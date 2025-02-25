package main

import (
	"context"
	//"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KarlHavoc/aggregatorCLI/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("useage: %s <name>", cmd.Name)
	}
	name := cmd.Arguments[0]
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		log.Fatal(err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %v", name)
	}
	fmt.Println("User has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("please enter a name")
	}
	name := cmd.Arguments[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name})
	if err != nil {
		log.Fatal(err)
	}
	s.cfg.SetUser(name)
	fmt.Printf("user %s was created\n", name)
	log.Printf("user: %v\ncreated_at: %v\nupdated_at: %v\nname: %v\n", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		log.Println("failed to delete users table")
	}
	fmt.Println("deleted all users from db")

	os.Exit(0)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if len(users) == 0 {
		fmt.Println("No users registered\nUse the command | register <user_name> | to register a new user")
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("%s (current)\n", user.Name)
			continue
		}
		fmt.Println(user.Name)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	var user_id uuid.UUID
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			user_id = user.ID
		}
	}

	if len(cmd.Arguments) != 2 {
		fmt.Println("please input a valid name and url")
		os.Exit(1)
	}
	name := cmd.Arguments[0]
	url := cmd.Arguments[1]

	new_feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user_id,
	})
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollows(context.Background(), database.CreateFeedFollowsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    new_feed.UserID,
		FeedID:    new_feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Print(new_feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		name, err := s.db.GetUserName(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Feed name: %v\n", feed.Name)
		fmt.Printf("Feed url: %v\n", feed.Url)
		fmt.Printf("Added by: %v\n", name)

	}
	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.Arguments) != 1 {
		log.Fatalf("useage: %v <url_to_follow>", cmd.Name)
	}
	url_to_follow := cmd.Arguments[0]
	current_user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	fmt.Println(current_user.Name)
	if err != nil {
		return err
	}
	feed_id, err := s.db.GetFeed(context.Background(), url_to_follow)
	if err != nil {
		return err
	}
	new_follow, err := s.db.CreateFeedFollows(context.Background(), database.CreateFeedFollowsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    current_user.ID,
		FeedID:    feed_id,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed name: %s\n", new_follow.FeedName)
	fmt.Printf("Current user: %s\n", current_user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	current_user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	feeds_following, err := s.db.GetFeedFollowsForUser(context.Background(), current_user.ID)
	if err != nil {
		return err
	}
	for _, feeds := range feeds_following {
		feed_name, err := s.db.GetFeedName(context.Background(), feeds.FeedID)
		if err != nil {
			return nil
		}
		fmt.Println("Following:")
		fmt.Println(feed_name)

	}
	return nil
}
