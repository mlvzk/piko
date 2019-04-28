package fourchan

import (
	"flag"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://boards.4channel.org"

var update = flag.Bool("update", false, "update .golden files")

func TestFourchanIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://boards.4channel.org/g/": true,
		"boards.4channel.org/g/":         true,
		"https://boards.4channel.org/g/thread/70377765/hpg-esg-headphone-general": true,
		"https://boards.4chan.org/pol/":                                           true,
		"https://imgur.com/":                                                      false,
	}

	for target, expected := range tests {
		if (Fourchan{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := FourchanIterator{
		url: ts.URL + "/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"title":        "1306532808724.png",
				"imgURL":       "https://i.4cdn.org/adv/1554743536847.png",
				"id":           "1554743536847",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554743536847s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
		service.Item{
			Meta: map[string]string{
				"title":        "1306532948201.png",
				"imgURL":       "https://i.4cdn.org/adv/1554745883162.png",
				"id":           "1554745883162",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554745883162s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
