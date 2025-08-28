package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dbfletcher/gator/internal/config"
)

// state holds the application's state, including the configuration.
type state struct {
	cfg *config.Config
}

// command represents a command to be executed by the CLI.
type command struct {
	name string
	args []string
}

// commands is a registry for all available CLI commands.
type commands struct {
	handlers map[string]func(*state, command) error
}

// register adds a new command handler to the registry.
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// run executes a registered command.
func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command not found: %s", cmd.name)
	}
	return handler(s, cmd)
}

// handlerLogin handles the "login" command. It sets the current user in the config.
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("a username is required")
	}
	username := cmd.args[0]
	err := s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("could not set user: %w", err)
	}
	fmt.Printf("User set to: %s\n", username)
	return nil
}

func main() {
	// Read the application configuration.
	cfg, err := config.Read()
	// If the config file doesn't exist, we can proceed with a default/empty config.
	// The SetUser call will create the file. We only fail if another error occurs.
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("error reading config: %v", err)
	}

	// Initialize the application state.
	appState := state{
		cfg: &cfg,
	}

	// Create a new command registry and initialize its handler map.
	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	// Register the "login" command handler.
	cmds.register("login", handlerLogin)

	// Get command-line arguments. The first argument is the program name.
	args := os.Args
	if len(args) < 2 {
		log.Fatal("not enough arguments, expected a command")
	}

	// Parse the command and its arguments from the command-line input.
	cmd := command{
		name: args[1],
		args: args[2:],
	}

	// Run the command.
	err = cmds.run(&appState, cmd)
	if err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
