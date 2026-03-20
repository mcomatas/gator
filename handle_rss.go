package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mcomatas/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Item 		[]RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: aggregate <time between requests> ex: 1s, 1m, 1h")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	ticker := time.NewTicker(timeBetweenRequests)
	fmt.Println("Collecting feeds every", timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.Args[0]
	feedURL := cmd.Args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: name,
		Url: feedURL,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	_, err = s.db.CreateFollowFeed(context.Background(), database.CreateFollowFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil

}

func handlerGetFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}
	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	follow_feed, err := s.db.CreateFollowFeed(context.Background(), database.CreateFollowFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println(follow_feed.FeedName)
	fmt.Println(follow_feed.UserName)
	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	err = s.db.DeleteFollowFeed(context.Background(), database.DeleteFollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

func handlerGetFeedFollows(s *state, cmd command, user database.User) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("no args expected")
	}

	follow_feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	for _, follow := range follow_feeds {
		fmt.Println(follow.FeedName)
	}
	return nil
}

func parsePublishedAt(pubDate string) sql.NullTime {
	if pubDate == "" {
		return sql.NullTime{Valid: false}
	}
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, format := range formats {
		t, err := time.Parse(format, pubDate)
		if err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	return sql.NullTime{Valid: false}
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println("no feeds to fetch: %w", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		fmt.Println("failed to mark feed fetched: %w", err)
		return
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Println("failed to get feed: %w", err)
		return
	}

	for _, item := range feed.Channel.Item {
		_, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: parsePublishedAt(item.PubDate),
			FeedID:      nextFeed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate") {
				continue
			}
			fmt.Println("failed to create post:", err)
		}
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.Args) == 1 {
		n, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(n)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: limit,
	})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Url)
		fmt.Println("---")
	}
	return nil
}
