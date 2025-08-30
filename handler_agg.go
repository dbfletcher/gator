package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Printf("couldn't get next feed to fetch: %v", err)
		return
	}

	log.Printf("Fetching feed '%s'...", feed.Name)
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("couldn't mark feed as fetched: %v", err)
		return
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Printf("couldn't fetch feed: %v", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		log.Printf("Found post: '%s'", item.Title)
	}
	log.Printf("Feed '%s' collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}

	durationStr := cmd.Args[0]
	timeBetweenRequests, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s...\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	// Run immediately, then on each tick
	scrapeFeeds(s)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
