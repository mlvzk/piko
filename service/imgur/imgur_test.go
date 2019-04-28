package imgur

import (
	"flag"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://imgur.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://imgur.com/gallery/kgIfZrm":     true,
		"imgur.com/gallery/kgIfZrm":             true,
		"https://imgur.com/t/article13/y2Vp0nZ": true,
		"https://youtube.com/":                  false,
	}

	for target, expected := range tests {
		if (Imgur{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := ImgurIterator{
		url: ts.URL + "/t/article13/EfY6CxU",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"id":         "o2nusiZ",
				"ext":        "png",
				"itemType":   "http://schema.org/ImageObject",
				"albumTitle": "Some advice for those of you in the EU",
			},
			DefaultName: "%[id].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
