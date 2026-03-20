package main

import (
	"fmt"
	"context"
	"time"
	"github.com/mcomatas/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login requires a username")
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Println("User switched successfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("register requires a username")
	}
	name := cmd.Args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: 	   name,
	})
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Println("User created successfully:")
	printUser(user)
	return nil
}

func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("reset does not take arguments")
	}

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("get-users does not take arguments")
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Println("* ", user.Name, "(current)")
		} else {
			fmt.Println("* ", user.Name)
		}
	}
	return nil
}
