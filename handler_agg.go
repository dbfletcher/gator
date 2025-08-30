package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dbfletcher/gator/internal/database"
	"github.com/google/uuid"
)

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("no feeds to fetch")
			return
		}
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
		// Handle different time formats
		publishedAt, err := parseTime(item.PubDate)
		if err != nil {
			log.Printf("couldn't parse publish date '%s' for post '%s': %v", item.PubDate, item.Title, err)
			continue
		}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: sql.NullTime{
				Time:  publishedAt,
				Valid: !publishedAt.IsZero(),
			},
			FeedID: feed.ID,
		})
		if err != nil {
			// Check for unique constraint violation
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("couldn't create post '%s': %v", item.Title, err)
		}
	}
	log.Printf("Feed '%s' collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}

func parseTime(t string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range layouts {
		parsedTime, err := time.Parse(layout, t)
		if err == nil {
			return parsedTime, nil
		}
	}
	return time.Time{}, fmt.Errorf("couldn't parse time '%s' with any known layout", t)
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
	go scrapeFeeds(s)
	for ; ; <-ticker.C {
		go scrapeFeeds(s)
	}
}

