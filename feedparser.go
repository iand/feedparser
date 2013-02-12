/*
  PUBLIC DOMAIN STATEMENT
  To the extent possible under law, Ian Davis has waived all copyright
  and related or neighboring rights to this Source Code file.
  This work is published from the United Kingdom.
*/

// Simple parser for RSS and Atom feeds
package feedparser

import (
	"encoding/xml"
	"io"
	"strings"
	"time"
)

type Feed struct {
	Title    string
	Subtitle string
	Link     string
	Items    []*FeedItem
}

type FeedItem struct {
	Id          string
	Title       string
	Description string
	Link        string
	When        time.Time
}

const feedTitle = "title"

const (
	rssChannel     = "channel"
	rssItem        = "item"
	rssLink        = "link"
	rssPubDate     = "pubdate"
	rssDescription = "description"
	rssId          = "guid"
)

const (
	atomSubtitle = "subtitle"
	atomFeed     = "feed"
	atomEntry    = "entry"
	atomLink     = "link"
	atomLinkHref = "href"
	atomUpdated  = "updated"
	atomSummary  = "summary"
	atomId       = "id"
)

const (
	levelFeed = 1
	levelPost = 2
)

func parseTime(f, v string) time.Time {
	t, err := time.Parse(f, v)
	if err != nil || v == "" {
		return time.Now()
	}
	return t
}

func NewFeed(r io.Reader) (*Feed, error) {
	var ns string
	var tag string
	var atom bool
	var level int
	feed := &Feed{}
	item := &FeedItem{}
	parser := xml.NewDecoder(r)
	for {
		token, err := parser.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			ns = strings.ToLower(t.Name.Space)
			tag = strings.ToLower(t.Name.Local)
			switch {
			case tag == atomFeed:
				atom = true
				level = levelFeed
			case tag == rssChannel:
				atom = false
				level = levelFeed
			case (!atom && tag == rssItem) || (atom && tag == atomEntry):
				level = levelPost
				item = &FeedItem{When: time.Now()}
			case atom && tag == atomLink:
				for _, a := range t.Attr {
					if strings.ToLower(a.Name.Local) == atomLinkHref {
						switch level {
						case levelFeed:
							feed.Link = a.Value
						case levelPost:
							item.Link = a.Value
						}
						break
					}
				}
			}
		case xml.EndElement:
			e := strings.ToLower(t.Name.Local)
			if e == atomEntry || e == rssItem {
				feed.Items = append(feed.Items, item)
			}
		case xml.CharData:
			text := string([]byte(t))
			if strings.TrimSpace(text) == "" {
				continue
			}
			switch level {
			case levelFeed:
				switch {
				case tag == feedTitle:
					feed.Title = text
				case (!atom && tag == rssDescription) || (atom && tag == atomSubtitle):
					feed.Subtitle = text
				case !atom && tag == rssLink:
					feed.Link = text
				}
			case levelPost:
				switch {
				case (!atom && tag == rssId) || (atom && tag == atomId):
					item.Id = text
				case ns == "" && tag == feedTitle:
					item.Title = text
				case (!atom && tag == rssDescription) || (atom && tag == atomSummary):
					item.Description = text
				case !atom && tag == rssLink:
					item.Link = text
				case atom && tag == atomUpdated:
					var f string
					switch {
					case strings.HasSuffix(strings.ToUpper(text), "Z"):
						f = "2006-01-02T15:04:05Z"
					default:
						f = "2006-01-02T15:04:05-07:00"
					}
					item.When = parseTime(f, text)
				case !atom && tag == rssPubDate:
					var f string
					if strings.HasSuffix(strings.ToUpper(text), "T") {
						f = "Mon, 2 Jan 2006 15:04:05 MST"
					} else {
						f = "Mon, 2 Jan 2006 15:04:05 -0700"
					}
					item.When = parseTime(f, text)
				}
			}
		}
	}
	return feed, nil
}
