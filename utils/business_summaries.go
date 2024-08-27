package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ethanhosier/mia-backend-go/types"
)

func countSlashes(u *url.URL) int {
	return strings.Count(strings.Trim(u.Path, "/"), "/")
}

func SortURLsByProximity(urls []string) ([]string, error) {
	parsedURLs := make([]*url.URL, len(urls))
	for i, u := range urls {
		parsed, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL %q: %v", u, err)
		}
		parsedURLs[i] = parsed
	}

	sort.Slice(parsedURLs, func(i, j int) bool {
		return countSlashes(parsedURLs[i]) < countSlashes(parsedURLs[j])
	})

	sortedURLs := make([]string, len(urls))
	for i, u := range parsedURLs {
		sortedURLs[i] = u.String()
	}

	return sortedURLs, nil
}

func SortURLPairsByProximity(pairs []types.UrlHtmlPair) ([]types.UrlHtmlPair, error) {
	parsedURLs := make([]*url.URL, len(pairs))
	for i, pair := range pairs {
		parsed, err := url.Parse(pair.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL %q: %v", pair.Url, err)
		}
		parsedURLs[i] = parsed
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return countSlashes(parsedURLs[i]) < countSlashes(parsedURLs[j])
	})

	return pairs, nil
}

func ImageUrlsAndText(html string) ([]string, string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	if err != nil {
		return nil, "", err
	}

	imageUrls, err := extractImageUrls(html)

	var textBuilder strings.Builder
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		text := item.Text()
		text = CleanText(text)
		if text != "" {
			textBuilder.WriteString(text + "\n")
		}
	})

	return imageUrls, textBuilder.String(), err
}

func ExtractGeneralWebsiteData(html string) ([]string, string, error) {
	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return []string{}, "", err
	}

	// Extract data
	data := types.WebsiteData{
		Title:    doc.Find("title").Text(),
		Headings: make(map[string][]string),
	}

	// Meta description
	data.MetaDescription, _ = doc.Find("meta[name='description']").Attr("content")
	data.MetaDescription = CleanText(data.MetaDescription)

	// Headings (H1, H2, etc.)
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		data.Headings[tag] = append(data.Headings[tag], CleanText(s.Text()))
	})

	// Keywords
	data.Keywords, _ = doc.Find("meta[name='keywords']").Attr("content")
	data.Keywords = CleanText(data.Keywords)

	// Extract the first 5 paragraphs
	var paragraphs []string
	doc.Find("p").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i < 5 {
			paragraphs = append(paragraphs, CleanText(s.Text()))
			return true
		}
		return false
	})
	data.Summary = strings.Join(paragraphs, "\n\n") // Combine paragraphs with double line breaks for readability

	// Convert to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return []string{}, "", err
	}

	imageUrls, err := extractImageUrls(html)

	return imageUrls, string(jsonData), err
}

func extractImageUrls(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var imageUrls []string

	doc.Find("img").Each(func(index int, item *goquery.Selection) {
		if src, exists := item.Attr("src"); exists {
			imageUrls = append(imageUrls, src)
		}
	})

	doc.Find("*").Each(func(index int, item *goquery.Selection) {
		style, exists := item.Attr("style")
		if exists {
			re := regexp.MustCompile(`background(?:-image)?\s*:\s*url\(['"]?([^'")]+)['"]?\)`)
			matches := re.FindAllStringSubmatch(style, -1)
			for _, match := range matches {
				if len(match) > 1 {
					imageUrls = append(imageUrls, match[1])
				}
			}
		}
	})

	return imageUrls, nil
}
