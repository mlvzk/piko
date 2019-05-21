// Copyright 2019 mlvzk
// This file is part of the piko library.
//
// The piko library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The piko library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the piko library. If not, see <http://www.gnu.org/licenses/>.

package soundcloud

import (
	"flag"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const baseApiURL = "https://api.soundcloud.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator": true,
		"https://soundcloud.com/":              true,
		"https://soundcloud.com/search?q=test": true,
		"https://soundcloud.com/fadermedia":    true,
		"https://instagram.com/":               false,
	}

	for target, expected := range tests {
		if (Soundcloud{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, baseApiURL, *update)
	defer ts.Close()

	iterator := SoundcloudIterator{
		url:        "https://soundcloud.com/ishaan-bhagwakar/oldie-ofwgkta",
		baseApiURL: ts.URL,
		clientID:   "a3e059563d7fd3372b49b37f00a00bcf",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	for _, item := range items {
		item.Meta["playCount"] = "ignore"
		item.Meta["_downloadURL"] = "ignore"
	}

	expected := []service.Item{
		{
			Meta: map[string]string{
				"id":           "224754696",
				"title":        "Oldie - OFWGKTA",
				"username":     "Ishaan Bhagwakar",
				"createdAt":    "2015/09/20 17:32:13 +0000",
				"duration":     "636453",
				"playCount":    "ignore",
				"ext":          "mp3",
				"_downloadURL": "ignore",
			},
			DefaultName: "%[title].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
