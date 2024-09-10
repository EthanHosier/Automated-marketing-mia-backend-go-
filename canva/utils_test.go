package canva

import (
	"bytes"
	"image/color"
	"testing"
)

func colorEqual(c1, c2 color.Color) bool {
	if c1 == nil || c2 == nil {
		return c1 == c2
	}

	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		hex      string
		expected color.Color
		err      bool
	}{
		{"#FF5733", color.RGBA{R: 255, G: 87, B: 51, A: 255}, false},
		{"FF5733", color.RGBA{R: 255, G: 87, B: 51, A: 255}, false},
		{"#000000", color.RGBA{R: 0, G: 0, B: 0, A: 255}, false},
		{"#FFFFFF", color.RGBA{R: 255, G: 255, B: 255, A: 255}, false},
		{"#ABC", nil, true},    // Invalid length
		{"#GGGGGG", nil, true}, // Invalid hex digits
	}

	for _, tt := range tests {
		got, err := hexToColor(tt.hex)
		if (err != nil) != tt.err {
			t.Errorf("hexToColor(%v) error = %v, wantErr %v", tt.hex, err, tt.err)
			continue
		}
		if !colorEqual(got, tt.expected) {
			t.Errorf("hexToColor(%v) = %v, want %v", tt.hex, got, tt.expected)
		}
	}
}

func TestCreateColorImage(t *testing.T) {
	tests := []struct {
		hex      string
		expected []byte
		err      bool
	}{
		{"#FF5733", generateExpectedImageData("#FF5733"), false},
		{"#000000", generateExpectedImageData("#000000"), false},
		{"#FFFFFF", generateExpectedImageData("#FFFFFF"), false},
		{"#GGGGGG", nil, true},
	}

	for _, tt := range tests {
		got, err := createColorImage(tt.hex)
		if (err != nil) != tt.err {
			t.Errorf("createColorImage(%v) error = %v, wantErr %v", tt.hex, err, tt.err)
			continue
		}

		if err == nil && !bytes.Equal(got, tt.expected) {
			t.Errorf("createColorImage(%v) = %v, want %v", tt.hex, got, tt.expected)
		}
	}
}

// Helper function to generate expected image data for known inputs
func generateExpectedImageData(hexColor string) []byte {
	img, err := createColorImage(hexColor)
	if err != nil {
		return nil
	}
	return img
}

func TestPopulateTemplateInputData(t *testing.T) {
	imageFields := []ImageField{
		{Name: "logo", AssetId: "123"},
	}
	textFields := []TextField{
		{Name: "headline", Text: "Welcome"},
	}
	colorFields := []ColorField{
		{Name: "background", ColorAssetId: "456"},
	}

	expected := map[string]map[string]string{
		"logo": {
			"type":     "image",
			"asset_id": "123",
		},
		"headline": {
			"type": "text",
			"text": "Welcome",
		},
		"background": {
			"type":     "image",
			"asset_id": "456",
		},
	}

	got := populateTemplateInputData(imageFields, textFields, colorFields)

	for key, expectedVal := range expected {
		gotVal, ok := got[key]
		if !ok {
			t.Errorf("populateTemplateInputData() missing key %v", key)
			continue
		}
		for subKey, subVal := range expectedVal {
			if gotVal[subKey] != subVal {
				t.Errorf("populateTemplateInputData() key %v, field %v: got %v, want %v", key, subKey, gotVal[subKey], subVal)
			}
		}
	}
}
