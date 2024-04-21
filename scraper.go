package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rushilp164/go_rss_demo/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func startScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Collecting feeds every %s on %v goroutines...", timeBetweenRequest, concurrency)
	ticker := time.NewTicker(timeBetweenRequest)

	// Que: Surprise again!!!!
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("Couldn't get next feeds to fetch", err)
			continue
		}

		log.Printf("Found %v feeds to fetch!", len(feeds))

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			log.Println(feed.Name)
			wg.Add(1)
			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}

}

// ToDo: Make a batches of x number of go routines for hit sql queries
func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	log.Println("Starting with feed:", feed.Name)

	feedData, err := fetchFeed(feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}

	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		// Que: what the heck is this?
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}
		_, err := db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			PublishedAt: publishedAt,
			Description: sql.NullString{String: item.Description, Valid: true},
			FeedID:      feed.ID,
		})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				continue
			}
			log.Printf("Couldn't create feed: %v", err)
			continue
		}
	}
	/** After everything mark feed as fetched
	If err, just printing the log */
	_, err = db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func fetchFeed(url string) (*RSSFeed, error) {
	httpClient := http.Client{}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed
	err = xml.Unmarshal(data, &rssFeed)
	if err != nil {
		return nil, err
	}

	return &rssFeed, nil
}
