package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/ethanhosier/mia-backend-go/prompts"
	"github.com/ethanhosier/mia-backend-go/types"
	"github.com/sashabaranov/go-openai"
)

func PageScreenshot(url string) (string, error) {

	resp, err := http.Get(ScreenshotUrl + "?url=" + url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response types.ScreenshotScraperResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.ScreenshotBase64, nil
}

func Sitemap(url string, timeout int) ([]string, error) {
	resp, err := http.Get(SitemapScraperUrl + "?url=" + url + "&timeout=" + fmt.Sprintf("%d", timeout))
	if err != nil {
		return []string{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		return nil, err
	}

	// Filter out non-URL websites (e.g., ending in .xml, .pdf)
	var filteredUrls []string
	for _, u := range urls {
		if !strings.HasSuffix(u, ".xml") && !strings.HasSuffix(u, ".pdf") {
			filteredUrls = append(filteredUrls, u)
		}
	}

	return filteredUrls, nil
}

func ScrapeSinglePage(url string) (string, error) {
	resp, err := http.Get(SinglePageHtmlScraperUrl + "?url=" + url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func BusinessPageSummaries(url string, timeout int, llmClient *LLMClient) ([]string, error) {
	pages, err := scrapedPages(url, timeout)

	if err != nil {
		log.Println("Error getting scraped pages:", err, ". Trying again (1st retry)")
		pages, err = scrapedPages(url, timeout)

		if err != nil {
			log.Println("Error getting scraped pages:", err, ". Trying again (2nd retry)")
			pages, err = scrapedPages(url, timeout)

			if err != nil {
				return nil, fmt.Errorf("tried 3 times to get scraped business pages, failed last time with error: %v", err)
			}
		}
	}
	n := len(pages)

	log.Println("Successfully got data for", n, "scraped pages")

	wg := sync.WaitGroup{}
	wg.Add(n)

	ch := make(chan string, n)

	for _, page := range pages {
		go func(page string) {
			defer wg.Done()
			summary, err := llmClient.LlamaSummarise(prompts.ScrapedWebPageSummary+page, 200)

			if err != nil {
				log.Println("Error getting llama summary:", err)
			}
			ch <- summary
		}(page)
	}

	wg.Wait()
	close(ch)

	summaries := []string{}
	for summary := range ch {
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func scrapedPages(url string, timeout int) ([]string, error) {
	resp, err := http.Get(BusinessScraperUrl + "?url=" + url + "&timeout=" + fmt.Sprintf("%d", timeout))
	if err != nil {
		return []string{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scrapedPages []string
	err = json.Unmarshal(body, &scrapedPages)

	return scrapedPages, err
}

func BusinessSummaryPoints(jsonString string, llmClient *LLMClient) (*types.BusinessSummary, error) {
	completion, err := llmClient.OpenaiCompletion(prompts.BusinessSummary+jsonString, openai.GPT4o)

	if err != nil {
		return nil, err
	}

	var businessSummary types.BusinessSummary

	extractedObj := ExtractJsonObj(completion, CurlyBracket)

	err = json.Unmarshal([]byte(extractedObj), &businessSummary)

	return &businessSummary, err
}

func PageContentsScrape(url string) (*types.BodyContentsScrapeResponse, error) {
	resp, err := http.Get(SinglePageContentScraperUrl + url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response types.BodyContentsScrapeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
