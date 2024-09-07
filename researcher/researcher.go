package researcher

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ethanhosier/mia-backend-go/openai"
	"github.com/ethanhosier/mia-backend-go/services"
)

const maxBusinessSummaryUrls = 40

type Researcher struct {
	servicesClient *services.ServicesClient
	openaiClient   *openai.OpenaiClient
}

func New(sc *services.ServicesClient, oc *openai.OpenaiClient) *Researcher {

	return &Researcher{
		servicesClient: sc,
	}
}

func (r *Researcher) Sitemap(url string, timeout int) ([]string, error) {
	urls, err := r.servicesClient.Sitemap(url, timeout)
	if err != nil {
		return nil, err
	}

	sitemap := removeDuplicates(urls)
	return sitemap, nil
}

func (r *Researcher) BusinessSummary(url string) ([]string, *BusinessSummary, []string, error) {
	urls, err := r.Sitemap(url, 15)
	if err != nil {
		return nil, nil, nil, err
	}

	colors, err := r.ColorsFromUrl(url)
	if err != nil {
		return nil, nil, nil, err
	}

	sortedUrls, err := sortURLsByProximity(urls)
	if err != nil {
		return nil, nil, nil, err
	}

	imageUrls, bodyTexts, err := r.scrapeWebsitePages(sortedUrls[:min(maxBusinessSummaryUrls, len(sortedUrls))])

	if err != nil {
		return nil, nil, nil, err
	}

	jsonTexts, err := json.Marshal(bodyTexts)
	if err != nil {
		return nil, nil, nil, err
	}

	businessSummaries, err := r.businessSummaryPoints(string(jsonTexts))
	if err != nil {
		return nil, nil, nil, err
	}

	businessSummaries.Colors = colors

	return urls, businessSummaries, imageUrls, nil
}

func (r *Researcher) ColorsFromUrl(url string) ([]string, error) {
	screenshotBase64, err := r.servicesClient.PageScreenshot(url)
	if err != nil {
		return nil, fmt.Errorf("error taking screenshot of page: %v", err)
	}

	resp, err := r.openaiClient.ImageCompletion(context.TODO(), openai.ColorThemesPrompt, []string{screenshotBase64}, openai.GPT4o)
	if err != nil {
		return nil, err
	}

	var colors []string
	err = json.Unmarshal([]byte(resp), &colors)
	if err != nil {
		return nil, err
	}

	return colors, nil
}

func (r *Researcher) scrapeWebsitePages(urls []string) ([]string, []string, error) {
	n := len(urls)

	pageWg := sync.WaitGroup{}
	pageWg.Add(n)

	pageCh := make(chan services.BodyContentsScrapeResponse, n)
	errorCh := make(chan error, n)

	for _, url := range urls {
		go func(url string) {
			defer pageWg.Done()

			pageContents, err := r.servicesClient.PageContentsScrape(url)
			if err != nil {
				errorCh <- err
				return
			}
			pageCh <- *pageContents
		}(url)
	}
	pageWg.Wait()
	close(pageCh)

	select {
	case err := <-errorCh:
		return nil, nil, err
	default:
	}

	imageSet := make(map[string]struct{})
	pageContents := []string{}

	for contents := range pageCh {
		for _, imageUrl := range contents.ImageUrls {
			imageSet[imageUrl] = struct{}{}
		}
		pageContents = append(pageContents, contents.Contents.String())
	}

	images := make([]string, 0, len(imageSet))
	for imageUrl := range imageSet {
		images = append(images, imageUrl)
	}

	return images, pageContents, nil
}

func (r *Researcher) businessSummaryPoints(jsonString string) (*BusinessSummary, error) {
	completion, err := r.openaiClient.ChatCompletion(context.TODO(), openai.BusinessSummaryPrompt+jsonString, openai.GPT4o)

	if err != nil {
		return nil, err
	}

	var businessSummary BusinessSummary

	extractedObj := openai.ExtractJsonData(completion, openai.JSONObj)
	err = json.Unmarshal([]byte(extractedObj), &businessSummary)

	return &businessSummary, err
}
