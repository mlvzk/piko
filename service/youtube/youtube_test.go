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

package youtube

import (
	"flag"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/piko/service/testutil"
)

const base = "https://www.youtube.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://www.youtube.com/watch?v=HOK0uF-Z0xM": true,
		"https://youtube.com/watch?v=HOK0uF-Z0xM":     true,
		"youtube.com/watch?v=HOK0uF-Z0xM":             true,
		"https://youtu.be/HOK0uF-Z0xM":                true,
		"https://imgur.com/":                          false,
	}

	for target, expected := range tests {
		if (Youtube{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := testutil.CacheHttpRequest(t, base, *update)
	defer ts.Close()

	iterator := YoutubeIterator{
		urls: []string{ts.URL + "/watch?v=Q8Tiz6INF7I"},
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	for k := range items[0].Meta {
		if k[0] == '_' {
			items[0].Meta[k] = "ignore"
		}
	}

	expected := []service.Item{
		{
			Meta: map[string]string{
				"_ytConfig": "ignore",
				"author":    "Andres Trevino",
				"title":     "Hit the road Jack!",
				"ext":       "mkv",
			},
			DefaultName: "%[title].%[ext]",
			AvailableOptions: map[string]([]string){
				"quality":   []string{"best", "medium", "worst"},
				"useFfmpeg": []string{"yes", "no"},
				"onlyAudio": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"quality":   "medium",
				"useFfmpeg": "yes",
				"onlyAudio": "no",
			},
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}

func TestExtractURLs(t *testing.T) {
	tests := map[string]struct {
		target    string
		wantLinks []string
		wantErr   bool
	}{
		"short playlist": {
			target: "/playlist?list=PLE2y3n8EQ6Vwnua4Dhfm6lI9zoy2IzDOD",
			wantLinks: []string{
				"https://www.youtube.com/watch?v=ycxZPpmaxxs",
				"https://www.youtube.com/watch?v=bALlZMHqy3g",
				"https://www.youtube.com/watch?v=UsSf9zFfgsU",
				"https://www.youtube.com/watch?v=SLAmSo0Gd4U",
				"https://www.youtube.com/watch?v=6VhhqeF3V_4",
				"https://www.youtube.com/watch?v=DIy1YnhvRLY",
				"https://www.youtube.com/watch?v=2nHoZZ-P8bI",
				"https://www.youtube.com/watch?v=Gibx29aqtcM",
				"https://www.youtube.com/watch?v=7AD9r0uUASE",
				"https://www.youtube.com/watch?v=ICqjNK1bDqw",
				"https://www.youtube.com/watch?v=wUq416NjXbk",
				"https://www.youtube.com/watch?v=Usfizu445gE",
				"https://www.youtube.com/watch?v=NnGrlVDO8yU",
				"https://www.youtube.com/watch?v=0b0Xj-eVneY",
				"https://www.youtube.com/watch?v=jx9DLjqngVs",
				"https://www.youtube.com/watch?v=lS4hEl_7BbI",
			},
		},
		"long playlist over 100 videos": {
			target: "/playlist?list=PLvn31dJXvzXpsspKfYLuhqeyRbMtMu0wL",
			wantLinks: []string{
				"https://www.youtube.com/watch?v=05qcA4KPI0k",
				"https://www.youtube.com/watch?v=18uuczwHp78",
				"https://www.youtube.com/watch?v=1oZPtbXBn-4",
				"https://www.youtube.com/watch?v=1t-gK-9EIq4",
				"https://www.youtube.com/watch?v=2e6OoeY1X8c",
				"https://www.youtube.com/watch?v=3yg0-e5ZFqY",
				"https://www.youtube.com/watch?v=4dBtfeoXM8I",
				"https://www.youtube.com/watch?v=6GfkQOhXg_M",
				"https://www.youtube.com/watch?v=75nKyI2FFFI",
				"https://www.youtube.com/watch?v=7dgrMSTalZ0",
				"https://www.youtube.com/watch?v=8Bv802FwvCY",
				"https://www.youtube.com/watch?v=8yn3ViE6mhY",
				"https://www.youtube.com/watch?v=9Y5eqpVQ1p8",
				"https://www.youtube.com/watch?v=9pt7EWFF_T8",
				"https://www.youtube.com/watch?v=AZRGPg5laDU",
				"https://www.youtube.com/watch?v=A_p-myZaodg",
				"https://www.youtube.com/watch?v=BOrnC3LQeLs",
				"https://www.youtube.com/watch?v=B_geuq76Cig",
				"https://www.youtube.com/watch?v=BaBR--4bw08",
				"https://www.youtube.com/watch?v=C4kVQnZhHmg",
				"https://www.youtube.com/watch?v=CnA0ft6DMpM",
				"https://www.youtube.com/watch?v=DRPi0XXmc-I",
				"https://www.youtube.com/watch?v=FTdcEoBLEuk",
				"https://www.youtube.com/watch?v=FWRfpC8s6XU",
				"https://www.youtube.com/watch?v=Fy7FzXLin7o",
				"https://www.youtube.com/watch?v=GrC_yuzO-Ss",
				"https://www.youtube.com/watch?v=HBBFufxHj3M",
				"https://www.youtube.com/watch?v=IUWYPbe96jE",
				"https://www.youtube.com/watch?v=I_O37cE1j64",
				"https://www.youtube.com/watch?v=IsvfofcIE1Q",
				"https://www.youtube.com/watch?v=JIrm0dHbCDU",
				"https://www.youtube.com/watch?v=JPb-59BPHAk",
				"https://www.youtube.com/watch?v=KANBeat7FFA",
				"https://www.youtube.com/watch?v=KCRDQ2qwnds",
				"https://www.youtube.com/watch?v=KEoU0pgnFNc",
				"https://www.youtube.com/watch?v=M2VBmHOYpV8",
				"https://www.youtube.com/watch?v=MzGnX-MbYE4",
				"https://www.youtube.com/watch?v=NRJh_r1LiqY",
				"https://www.youtube.com/watch?v=NihMVuspKQw",
				"https://www.youtube.com/watch?v=OL8Wqe-QWM8",
				"https://www.youtube.com/watch?v=OSjxK1SrCWk",
				"https://www.youtube.com/watch?v=SsKyxkfj8ak",
				"https://www.youtube.com/watch?v=TPqLJJfrVVY",
				"https://www.youtube.com/watch?v=U9vfK_bl4o8",
				"https://www.youtube.com/watch?v=UgTl4wLlMGI",
				"https://www.youtube.com/watch?v=V7GCrTFCXYo",
				"https://www.youtube.com/watch?v=VEAuMiKqP-4",
				"https://www.youtube.com/watch?v=VkqXIpl7a2w",
				"https://www.youtube.com/watch?v=WAXfhWUFIGA",
				"https://www.youtube.com/watch?v=WWJem7RuBpc",
				"https://www.youtube.com/watch?v=XWK7QLvuI-I",
				"https://www.youtube.com/watch?v=XkB4COqwcW4",
				"https://www.youtube.com/watch?v=Z3U8I0Bktb4",
				"https://www.youtube.com/watch?v=Z62fegq1gkk",
				"https://www.youtube.com/watch?v=ZUWvXERYJfk",
				"https://www.youtube.com/watch?v=_-QPvffO1gs",
				"https://www.youtube.com/watch?v=_1JAwLrQy9k",
				"https://www.youtube.com/watch?v=_6FBfAQ-NDE",
				"https://www.youtube.com/watch?v=a46z0mS3NPM",
				"https://www.youtube.com/watch?v=a8gYRf3aeQc",
				"https://www.youtube.com/watch?v=aDgHXiWgKlE",
				"https://www.youtube.com/watch?v=aGSKrC7dGcY",
				"https://www.youtube.com/watch?v=b1Wvvk4YtmE",
				"https://www.youtube.com/watch?v=bt-28iNQnwY",
				"https://www.youtube.com/watch?v=cGvZyrhObrg",
				"https://www.youtube.com/watch?v=cfzAGk8SlfE",
				"https://www.youtube.com/watch?v=dKnjm5SJ5jc",
				"https://www.youtube.com/watch?v=du8JSARa1H8",
				"https://www.youtube.com/watch?v=ejQ7KxUeItY",
				"https://www.youtube.com/watch?v=euBr4iyY_x8",
				"https://www.youtube.com/watch?v=f95pB9spuFk",
				"https://www.youtube.com/watch?v=fphsbLtrDe8",
				"https://www.youtube.com/watch?v=h1mD-_DKHc0",
				"https://www.youtube.com/watch?v=hUun8wjHx5Y",
				"https://www.youtube.com/watch?v=iDoSbyGBmy4",
				"https://www.youtube.com/watch?v=iEH4eqtK8SU",
				"https://www.youtube.com/watch?v=iTKJ_itifQg",
				"https://www.youtube.com/watch?v=j7EsBK4Mr80",
				"https://www.youtube.com/watch?v=jsCR05oKROA",
				"https://www.youtube.com/watch?v=kqRGZtGNPW4",
				"https://www.youtube.com/watch?v=l35XzUD8GGU",
				"https://www.youtube.com/watch?v=lD87Hbm9mrI",
				"https://www.youtube.com/watch?v=mU3tlDMI8xw",
				"https://www.youtube.com/watch?v=nhZdL4JlnxI",
				"https://www.youtube.com/watch?v=oeBTsGkngj8",
				"https://www.youtube.com/watch?v=pO0A998XZ5k",
				"https://www.youtube.com/watch?v=qU8UfYdKHvs",
				"https://www.youtube.com/watch?v=r_0sL_SQYvw",
				"https://www.youtube.com/watch?v=rxv9TTmk18o",
				"https://www.youtube.com/watch?v=snILjFUkk_A",
				"https://www.youtube.com/watch?v=u1xrNaTO1bI",
				"https://www.youtube.com/watch?v=up3r4qRWxWE",
				"https://www.youtube.com/watch?v=urbmwI8APdo",
				"https://www.youtube.com/watch?v=vOtRZOlE0WM",
				"https://www.youtube.com/watch?v=vXfbnS_BybQ",
				"https://www.youtube.com/watch?v=vyrpRzdvp5U",
				"https://www.youtube.com/watch?v=wkKueyJaA0A",
				"https://www.youtube.com/watch?v=yaGKZsgA_u0",
				"https://www.youtube.com/watch?v=zZeRwuN68VQ",
				"https://www.youtube.com/watch?v=zzbzHtdCzlI",
			},
		},
	}
	for ttName, tt := range tests {
		t.Run(ttName, func(t *testing.T) {
			ts := testutil.CacheHttpRequest(t, base, *update)
			defer ts.Close()

			gotLinks, err := Youtube{}.extractURLs(ts.URL + tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Youtube.ExtractURLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Strings(gotLinks)
			sort.Strings(tt.wantLinks)
			if diff := pretty.Compare(gotLinks, tt.wantLinks); diff != "" {
				t.Errorf("%s diff:\n%s", t.Name(), diff)
			}
		})
	}
}
