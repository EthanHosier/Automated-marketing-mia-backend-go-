package canva

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"
)

func hexToColor(hex string) (color.Color, error) {
	if strings.HasPrefix(hex, "#") {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color format")
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func createColorImage(hexColor string) ([]byte, error) {
	c, err := hexToColor(hexColor)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func populateTemplateInputData(imageFields []ImageField, textFields []TextField, colorFields []ColorField) map[string]map[string]string {
	inputData := map[string]map[string]string{}

	for _, field := range imageFields {
		inputData[field.Name] = map[string]string{
			"type":     "image",
			"asset_id": field.AssetId,
		}
	}

	for _, field := range textFields {
		inputData[field.Name] = map[string]string{
			"type": "text",
			"text": field.Text,
		}
	}

	for _, field := range colorFields {
		inputData[field.Name] = map[string]string{
			"type":     "image",
			"asset_id": field.ColorAssetId,
		}
	}

	return inputData
}
