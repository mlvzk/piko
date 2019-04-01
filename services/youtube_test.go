package shovel

import "testing"

func TestYoutubeIsValidTarget(t *testing.T) {
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
