# gator

A multi-user RSS feed aggregator CLI built in Go. Gator lets you add RSS feeds, follow them, and browse the latest posts — all from your terminal.

## Prerequisites

- [Go](https://golang.org/dl/) 1.22+
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

```bash
go install github.com/mcomatas/gator@latest
```

## Configuration

Create a config file at `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable"
}
```

Replace `<username>` and `<password>` with your Postgres credentials. Make sure you've created the `gator` database and run the migrations:

```bash
goose -dir sql/schema postgres "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable" up
```

## Commands

### User management

```bash
gator register <username>   # Create a new user and log in
gator login <username>      # Switch to an existing user
gator users                 # List all users
gator reset                 # Delete all users (and their data)
```

### Feeds

```bash
gator addfeed <name> <url>  # Add a new feed and follow it
gator feeds                 # List all feeds
gator follow <url>          # Follow an existing feed
gator unfollow <url>        # Unfollow a feed
gator following             # List feeds you're following
```

### Aggregation & browsing

```bash
gator agg <interval>        # Start the feed aggregator (e.g. 1m, 30s, 1h)
gator browse [limit]        # Browse latest posts (default limit: 2)
```

### Example workflow

```bash
# Register a user
gator register michael

# Add some feeds
gator addfeed "Boot.dev Blog" https://www.boot.dev/blog/index.xml
gator addfeed "Hacker News" https://news.ycombinator.com/rss
gator addfeed "TechCrunch" https://techcrunch.com/feed/

# Start aggregating in one terminal
gator agg 1m

# Browse posts in another terminal
gator browse 10
```
