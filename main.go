package main

import _ "github.com/lib/pq"

import (
	"github.com/mcomatas/gator/internal/config"
	"github.com/mcomatas/gator/internal/database"
	"database/sql"
	"log"
	"os"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	//fmt.Printf("Read config: %+v\n", cfg)

	db, err := sql.Open("postgres", cfg.DBURL)

	s := &state{
		cfg: &cfg,
		db:  database.New(db),
	}
	cmds := commands{registeredCommands: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAggregate)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollowFeed))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	cmds.register("following", middlewareLoggedIn(handlerGetFeedFollows))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := command{Name: os.Args[1], Args: os.Args[2:]}
	if err := cmds.run(s, cmd); err != nil {
		log.Fatal(err)
	}
}
