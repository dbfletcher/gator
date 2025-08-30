package main

import (
	"context"
	"fmt"

	"github.com/dbfletcher/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("you must be logged in to perform this action: %w", err)
		}
		return handler(s, cmd, user)
	}
}
