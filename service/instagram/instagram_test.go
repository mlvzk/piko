package instagram

import (
	"flag"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://www.instagram.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://www.instagram.com/explore/tags/cat/": true,
		"instagram.com/explore/tags/cat/":             true,
		"https://www.instagram.com/p/Bv3X1rVBWm5/":    true,
		"https://www.instagram.com/newding2/?hl=en":   true,
		"https://youtube.com/":                        false,
	}

	for target, expected := range tests {
		if (Instagram{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := InstagramIterator{
		url: ts.URL + "/p/BsOGulcndj-/",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	if correctURL := strings.Contains(items[0].Meta["imgURL"], "cdninstagram.com"); !correctURL {
		t.Fatalf("Incorrect imgURL")
	}
	items[0].Meta["imgURL"] = "ignore"

	expected := []service.Item{
		{
			Meta: map[string]string{
				"imgURL": "ignore",
				"caption": `Let’s set a world record together and get the most liked post on Instagram. Beating the current world record held by Kylie Jenner (18 million)! We got this 🙌

#LikeTheEgg #EggSoldiers #EggGang`,
				"author": "world_record_egg",
				"date":   "2019-01-04T17:05:45",
				"id":     "BsOGulcndj-",
				"ext":    "jpg",
			},
			DefaultName: "%[title].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
