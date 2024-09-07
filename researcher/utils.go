package researcher

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func sortURLsByProximity(urls []string) ([]string, error) {
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

func countSlashes(u *url.URL) int {
	return strings.Count(strings.Trim(u.Path, "/"), "/")
}

func removeDuplicates(urls []string) []string {
	uniqueUrls := make(map[string]bool)
	for _, url := range urls {
		uniqueUrls[url] = true
	}

	var result []string
	for url := range uniqueUrls {
		result = append(result, url)
	}

	return result
}
