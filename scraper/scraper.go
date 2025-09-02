package scraper

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
)

type ScraperInterface interface {
	Scrape(url string) (string, error)
}

type Scraper struct{}

func (s *Scraper) Scrape(url string) (string, error) {
	// Implementation goes here
	c := colly.NewCollector()
	scrapedText := ""

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting:", r.URL)
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		fmt.Println("Body found")
		// You can process the body content here
		scrapedText = e.Text
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Response received:", r.StatusCode)
	})

	log.Println("Starting to visit:", url)
	err := c.Visit(url)
	if err != nil {
		return "", err
	}

	return scrapedText, nil
}

func NewScraper() ScraperInterface {
	return &Scraper{}
}
