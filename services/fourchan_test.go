package shovel

import "testing"

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
			t.Errorf("Invalid result, expected: %v", expected)
		}
	}
}
