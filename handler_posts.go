package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/dbfletcher/gator/internal/database"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 10
	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			log.Printf("invalid limit '%s', using default of %d", cmd.Args[0], limit)
		} else {
			limit = parsedLimit
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Println("Recent posts from your feeds:")
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)
		if post.Description.Valid {
			fmt.Printf("Description: %s\n", post.Description.String)
		}
		if post.PublishedAt.Valid {
			fmt.Printf("Published At: %s\n", post.PublishedAt.Time.Format("2006-01-02 15:04:05"))
		}
		fmt.Println("---")
	}

	return nil
}
