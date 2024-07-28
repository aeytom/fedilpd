package main

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/aeytom/fedilpd/app"
	"github.com/mattn/go-mastodon"
	"github.com/mmcdole/gofeed"
)

var (
	tags = []string{
		"Berlin",
		"Polizei",
		"Friedrichshain",
		"Kreuzberg",
		"Pankow",
		"Charlottenburg",
		"Wilmersdorf",
		"Spandau",
		"Steglitz",
		"Zehlendorf",
		"Tempelhof",
		"Schöneberg",
		"Neukölln",
		"Treptow",
		"Köpenick",
		"Marzahn",
		"Hellersdorf",
		"Lichtenberg",
		"Reinickendorf",
	}
	tagsRe *regexp.Regexp
)

func main() {

	settings := app.LoadConfig()
	mc := settings.GetClient()

	tagsRe = regexp.MustCompile(`\b(` + strings.Join(tags, "|") + `)\b`)

	t := time.Now().AddDate(0, 0, -1)
	url := settings.Feed.Url + t.Format("02.01.2006")

	fp := gofeed.NewParser()
	fp.UserAgent = "fedi-lpd/0.2"
	resp, err := fp.ParseURL(url)
	if err != nil {
		settings.Fatal(err)
	}
	fmt.Println(resp.Title)

	for _, item := range resp.Items {
		fmt.Println(item.PublishedParsed.Format(time.RFC3339) + " " + item.Title + " " + item.Link)
		if !settings.StoreItem(item) {
			break
		}
	}

	for item := settings.GetUnsent(); item != nil; item = settings.GetUnsent() {
		title := hashtag(item.Title)
		link := regexp.MustCompile(`^.*\.(\d+)\.php$`).ReplaceAllString(item.Link, "https://berlin.de/-ii$1")
		footer := "\n\n" + hashtag(strings.Join(item.Categories, " ")) + "\n" + link
		status := hashtag(item.Description) + footer
		length := mblen(title + status)
		if length > 500 {
			status = hashtag(left(item.Description, 499-mblen(title)-mblen(footer))) + "…" + footer
		}
		toot := &mastodon.Toot{
			Status:      status,
			Sensitive:   true,
			SpoilerText: title,
			Visibility:  "public",
			Language:    "de",
			ScheduledAt: item.PublishedParsed,
		}
		if _, err := mc.PostStatus(context.Background(), toot); err != nil {
			settings.Logf("%s – %s – (%d/%d) :: %s", title, status, mblen(title), mblen(status), err.Error())
			settings.MarkError(item, err)
			continue
		} else {
			settings.MarkSent(item)
			settings.Log("… sent ", item.Link)
		}
	}

	replyNotifications(mc, settings)
}

func replyNotifications(mc *mastodon.Client, settings *app.Settings) {
	// Create a new Pagination object to control the number of notifications fetched
	pg := &mastodon.Pagination{
		MaxID:   "",
		SinceID: "",
		MinID:   "",
		Limit:   40,
	}
	// Fetch notifications using the Mastodon client
	if nl, err := mc.GetNotifications(context.Background(), pg); err != nil {
		// Log any errors that occur while fetching notifications
		settings.Log(err)
	} else {
		// Iterate over the fetched notifications
		for i, n := range nl {
			// Log the index, notification ID, type, and account name
			settings.Log(i, " ", n.ID, " ", n.Type, " ", n.Account.Acct)
			// Skip the notification if it's from a bot account
			if n.Account.Bot {
				continue
			}

			// Respond to different types of notifications with appropriate messages
			switch n.Type {
			// case "favourite":
			// 	sendReply(settings, mc, n, "Danke für ⭐") // Thank you for the star
			case "follow":
				sendReply(settings, mc, n, "Vielen Dank für das Interesse. 🤗") // Thank you for your interest
			// case "reblog":
			// 	sendReply(settings, mc, n, "Vielen Dank für die Unterstützung. 🤗") // Thank you for your support
			case "mention":
				doFavourite(settings, mc, n) // Favourite the mention
			}
		}
	}
	// Clear all notifications after processing
	if err := mc.ClearNotifications(context.Background()); err != nil {
		// Log any errors that occur while clearing notifications
		settings.Log(err)
	}
}

func doFavourite(settings *app.Settings, mc *mastodon.Client, note *mastodon.Notification) {
	settings.Log(note)
	if s, err := mc.Favourite(context.Background(), note.Status.ID); err != nil {
		settings.Log("doFavorite ", err)
	} else {
		settings.Log("doFavourite ", s.Account)
	}
}

func sendReply(settings *app.Settings, mc *mastodon.Client, note *mastodon.Notification, text string) {

	toot := &mastodon.Toot{
		Status:     "@" + note.Account.Acct + " " + text,
		Sensitive:  false,
		Visibility: "direct", // "unlisted"
		Language:   "de",
	}
	if note.Status != nil && note.Status.ID != "" {
		toot.InReplyToID = note.Status.ID
	}
	if _, err := mc.PostStatus(context.Background(), toot); err != nil {
		settings.Log(err)
	} else if err = mc.DismissNotification(context.Background(), note.ID); err != nil {
		settings.Log(err)
	}
}

func hashtag(text string) string {
	out := tagsRe.ReplaceAllString(html.UnescapeString(text), "#$1")
	return out
}

func mblen(text string) int {
	return len([]rune(text))
}

func left(input string, length int) string {
	asRunes := []rune(input)
	return string(asRunes[0 : length-1])
}
