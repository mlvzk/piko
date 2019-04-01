package shovel

import "testing"

func TestImgurIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://imgur.com/gallery/kgIfZrm":     true,
		"imgur.com/gallery/kgIfZrm":             true,
		"https://imgur.com/t/article13/y2Vp0nZ": true,
		"https://youtube.com/":                  false,
	}

	for target, expected := range tests {
		if (Imgur{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, expected: %v", expected)
		}
	}
}
