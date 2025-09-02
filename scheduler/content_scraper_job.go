package scheduler

import (
	"cine-pulse/model"
	"cine-pulse/notifier"
	"cine-pulse/scraper"
	"cine-pulse/storage"
	"context"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ContentScraperJob is a job that scrapes content and stores it in the database
type ContentScraperJob struct {
	scraper       scraper.ScraperInterface
	storage       *storage.SQLiteStorage
	modelMgr      *model.ModelManager
	sourceURLs    []string
	emailNotifier *notifier.EmailNotifier
	sendEmails    bool
}

// NewContentScraperJob creates a new content scraper job
func NewContentScraperJob(scraper scraper.ScraperInterface, storage *storage.SQLiteStorage, modelMgr *model.ModelManager, sourceURLs []string) *ContentScraperJob {
	// Get email configuration from environment variables
	emailConfig := notifier.GetEmailConfigFromEnv()
	var emailNotifier *notifier.EmailNotifier
	sendEmails := false

	// Only create email notifier if SMTP host and recipient are configured
	if emailConfig.SMTPHost != "" && emailConfig.RecipientEmail != "" {
		var err error
		emailNotifier, err = notifier.NewEmailNotifier(emailConfig)
		if err != nil {
			log.Printf("Failed to create email notifier: %v", err)
		} else {
			sendEmails = true
			log.Printf("Email notifications will be sent to: %s", emailConfig.RecipientEmail)
		}
	} else {
		log.Println("Email notifications disabled: missing configuration")
	}

	return &ContentScraperJob{
		scraper:       scraper,
		storage:       storage,
		modelMgr:      modelMgr,
		sourceURLs:    sourceURLs,
		emailNotifier: emailNotifier,
		sendEmails:    sendEmails,
	}
}

// Name returns the name of the job
func (j *ContentScraperJob) Name() string {
	return "content_scraper"
}

// Run executes the job
func (j *ContentScraperJob) Run(ctx context.Context) error {
	log.Printf("Running content scraper job with %d sources", len(j.sourceURLs))

	// If no sources are provided, use default
	if len(j.sourceURLs) == 0 {
		j.sourceURLs = []string{"https://nkiri.com/"}
	}

	var totalContentScraped int
	var allScrapedContent []storage.Content

	// Process each source URL
	for _, url := range j.sourceURLs {
		log.Printf("Scraping content from %s", url)

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Scrape the URL
		scrapedText, err := j.scraper.Scrape(url)
		if err != nil {
			log.Printf("Error scraping %s: %v", url, err)
			continue
		}

		// Set up the prompt for content extraction
		prompt := buildContentExtractionPrompt()

		// Try with Gemini model first
		geminiConfig := &model.ModelConfig{
			APIKey:    os.Getenv("GEMINI_API_KEY"),
			ModelName: "gemini-1.5-flash",
		}

		var response string
		var contents []storage.Content

		if geminiConfig.APIKey != "" {
			geminiModel, err := j.modelMgr.CreateModel(model.ModelTypeGemini, geminiConfig)
			if err != nil {
				log.Printf("Failed to create Gemini model: %v", err)
			} else {
				response, err = geminiModel.GenerateText(ctx, prompt+scrapedText)
				if err != nil {
					log.Printf("Error generating text with Gemini: %v", err)
				} else {
					// Manual JSON construction approach
					contents = extractContentManually(response)

					if len(contents) == 0 {
						// If manual extraction fails, fall back to standard approach
						processedResponse := preprocessModelResponse(response)

						// Parse response
						if err := json.Unmarshal([]byte(processedResponse), &contents); err != nil {
							log.Printf("Error parsing Gemini JSON response: %v", err)
							log.Printf("Raw response (first 100 chars): %s", truncateString(response, 100))
							log.Printf("Processed response (first 100 chars): %s", truncateString(processedResponse, 100))
							contents = nil
						}
					}
				}
			}
		}

		// If Gemini failed, try OpenAI
		if len(contents) == 0 && os.Getenv("OPENAI_API_KEY") != "" {
			openaiConfig := &model.ModelConfig{
				APIKey:    os.Getenv("OPENAI_API_KEY"),
				ModelName: "gpt-4o",
			}

			openaiModel, err := j.modelMgr.CreateModel(model.ModelTypeOpenAI, openaiConfig)
			if err != nil {
				log.Printf("Failed to create OpenAI model: %v", err)
			} else {
				response, err = openaiModel.GenerateText(ctx, prompt+scrapedText)
				if err != nil {
					log.Printf("Error generating text with OpenAI: %v", err)
				} else {
					// Preprocess response to extract valid JSON
					processedResponse := preprocessModelResponse(response)

					// Parse response
					if err := json.Unmarshal([]byte(processedResponse), &contents); err != nil {
						log.Printf("Error parsing OpenAI JSON response: %v", err)
						log.Printf("Raw response: %s", response)
						log.Printf("Processed response: %s", processedResponse)
						contents = nil
					}
				}
			}
		}

		// Save contents to database and collect for email notification
		if len(contents) > 0 {
			log.Printf("Extracted %d content items from %s", len(contents), url)

			// Add source URL to each content item
			sourceURL := url
			var scrapedContentForSource []storage.Content

			for i := range contents {
				contents[i].SourceURL = &sourceURL
				// The scraped_at timestamp will be set by the database
			}

			// Save to database
			for _, content := range contents {
				if err := j.storage.SaveContent(content); err != nil {
					log.Printf("Error saving content %s: %v", content.Title, err)
				} else {
					totalContentScraped++
					scrapedContentForSource = append(scrapedContentForSource, content)
				}
			}

			// Add successfully saved content to our collection for email
			allScrapedContent = append(allScrapedContent, scrapedContentForSource...)
		} else {
			log.Printf("No content extracted from %s", url)
		}
	}

	// Log job summary
	log.Printf("Content scraper job complete. Scraped %d content items from %d sources",
		totalContentScraped, len(j.sourceURLs))

	// Send email notification if content was scraped and email notifications are enabled
	if j.sendEmails && j.emailNotifier != nil && len(allScrapedContent) > 0 {
		log.Printf("Sending email notification with %d content items", len(allScrapedContent))
		if err := j.emailNotifier.NotifyContentUpdate(allScrapedContent, j.sourceURLs); err != nil {
			log.Printf("Failed to send email notification: %v", err)
		}
	} else if len(allScrapedContent) > 0 {
		log.Println("Email notifications disabled or no content scraped")
	}

	return nil
}

// preprocessModelResponse cleans and extracts valid JSON from model responses
func preprocessModelResponse(response string) string {
	log.Printf("Raw model response (first 200 chars): %s", truncateString(response, 200))

	// Remove markdown code block markers
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")

	// Find the first '[' and the last ']' to extract only the JSON array
	startIdx := strings.Index(response, "[")
	if startIdx == -1 {
		log.Println("No JSON array found in response")
		// Let's check if we have a valid JSON object instead
		objectStart := strings.Index(response, "{")
		if objectStart != -1 {
			log.Println("Found JSON object instead of array, wrapping in array")
			objectEnd := strings.LastIndex(response, "}")
			if objectEnd != -1 && objectEnd > objectStart {
				jsonObj := response[objectStart : objectEnd+1]
				return "[" + jsonObj + "]" // Wrap the single object in an array
			}
		}
		log.Println("Full response: " + response)
		return "[]" // Return empty array if no JSON found
	}

	endIdx := strings.LastIndex(response, "]")
	if endIdx == -1 || endIdx <= startIdx {
		log.Println("Invalid JSON array format in response")
		log.Println("Full response: " + response)
		return "[]" // Return empty array if invalid format
	}

	// Extract the JSON array part, ensuring we get the entire array
	jsonPart := response[startIdx : endIdx+1]
	log.Printf("Extracted JSON (first 200 chars): %s", truncateString(jsonPart, 200))

	// Remove any unexpected backticks that might be in the content
	jsonPart = strings.ReplaceAll(jsonPart, "`", "")

	// Fix common JSON formatting issues that Gemini tends to produce
	// 1. Fix spaces between field names and colons (e.g., "title :" -> "title:")
	jsonPart = regexp.MustCompile(`"([^"]+)" :`).ReplaceAllString(jsonPart, `"$1":`)

	// 2. Fix missing quotes around string values
	jsonPart = regexp.MustCompile(`: *([^"{}\[\],\d][^{}\[\],\s]*),`).ReplaceAllString(jsonPart, `:"$1",`)
	jsonPart = regexp.MustCompile(`: *([^"{}\[\],\d][^{}\[\],\s]*)$`).ReplaceAllString(jsonPart, `:"$1"`)

	// 3. Replace any escaped quotes that might cause issues
	jsonPart = strings.ReplaceAll(jsonPart, "\\\"", "\"")

	// 4. Clean up any control characters that might have slipped in
	jsonPart = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(jsonPart, "")

	// 5. Check trailing commas in arrays and objects
	jsonPart = regexp.MustCompile(`,\s*\}`).ReplaceAllString(jsonPart, `}`)
	jsonPart = regexp.MustCompile(`,\s*\]`).ReplaceAllString(jsonPart, `]`)

	// Do a manual check to see if the extracted JSON is valid
	var testJson []interface{}
	if err := json.Unmarshal([]byte(jsonPart), &testJson); err != nil {
		log.Printf("Extracted JSON is not valid: %v", err)

		// Try to fix common quotes issues in JSON manually
		jsonPart = strings.ReplaceAll(jsonPart, "\"\"", "\"") // Fix double quotes
		jsonPart = strings.ReplaceAll(jsonPart, "''", "'")    // Fix double single quotes
		jsonPart = strings.ReplaceAll(jsonPart, "…", "...")   // Fix ellipsis

		// Try again after fixes
		if err := json.Unmarshal([]byte(jsonPart), &testJson); err != nil {
			log.Printf("JSON is still invalid after fixes: %v", err)
			log.Println("Full JSON extract: " + jsonPart)

			// As a last resort, try to reformat the JSON properly
			jsonPart = reformatJSON(jsonPart)
			return jsonPart
		}
	}

	log.Printf("Successfully extracted valid JSON array with %d items", len(testJson))
	return jsonPart
}

// reformatJSON attempts to reformat malformed JSON into valid JSON
func reformatJSON(input string) string {
	// This is a simplified reformatter for common issues

	// First, ensure we have an array
	if !strings.HasPrefix(input, "[") || !strings.HasSuffix(input, "]") {
		log.Println("Input is not a proper JSON array, returning empty array")
		return "[]"
	}

	// Strip the outer brackets to work with the content
	content := strings.TrimSpace(input[1 : len(input)-1])

	// Split by objects - look for closing brace followed by comma
	parts := strings.Split(content, "},")

	// Last part won't have a comma, so it needs special handling
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if !strings.HasSuffix(lastPart, "}") {
			lastPart = lastPart + "}"
		}
		parts[len(parts)-1] = lastPart

		// Add closing brace to all other parts
		for i := 0; i < len(parts)-1; i++ {
			parts[i] = parts[i] + "}"
		}
	}

	// Process each object
	var validObjects []string
	for _, part := range parts {
		// Ensure it's an object
		trimmed := strings.TrimSpace(part)
		if !strings.HasPrefix(trimmed, "{") {
			trimmed = "{" + trimmed
		}
		if !strings.HasSuffix(trimmed, "}") {
			trimmed = trimmed + "}"
		}

		// Validate the object
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
			// It's valid, keep it
			validObjStr, _ := json.Marshal(obj)
			validObjects = append(validObjects, string(validObjStr))
		} else {
			log.Printf("Skipping invalid object: %s", trimmed)
		}
	}

	// If we have any valid objects, return them as an array
	if len(validObjects) > 0 {
		return "[" + strings.Join(validObjects, ",") + "]"
	}

	// If all else fails, return empty array
	return "[]"
}

// extractContentManually extracts content entries from the model response using regex
func extractContentManually(response string) []storage.Content {
	var results []storage.Content

	// Define regex pattern to match content entries with the expected fields
	titlePattern := regexp.MustCompile(`"title":\s*"([^"]+)"`)
	yearPattern := regexp.MustCompile(`"year":\s*(\d+)`)
	categoryPattern := regexp.MustCompile(`"category":\s*"([^"]+)"`)
	extraInfoPattern := regexp.MustCompile(`"extra_info":\s*"([^"]+)"`)
	typePattern := regexp.MustCompile(`"type":\s*"([^"]+)"`)
	ratingPattern := regexp.MustCompile(`"rating":\s*(\d+(?:\.\d+)?)`)

	// Find all object blocks in the response
	objectPattern := regexp.MustCompile(`\{[^{}]*\}`)
	objects := objectPattern.FindAllString(response, -1)

	log.Printf("Found %d potential content objects in response", len(objects))

	for _, obj := range objects {
		var content storage.Content
		var valid bool = true

		// Extract title (required)
		if titleMatches := titlePattern.FindStringSubmatch(obj); len(titleMatches) > 1 {
			content.Title = titleMatches[1]
		} else {
			valid = false
			continue // Skip if no title
		}

		// Extract year (optional)
		if yearMatches := yearPattern.FindStringSubmatch(obj); len(yearMatches) > 1 {
			if year, err := strconv.Atoi(yearMatches[1]); err == nil {
				content.Year = &year
			}
		}

		// Extract category (required)
		if categoryMatches := categoryPattern.FindStringSubmatch(obj); len(categoryMatches) > 1 {
			content.Category = categoryMatches[1]
			// Skip Korean content
			if content.Category == "Korean" {
				valid = false
				continue
			}
		} else {
			valid = false
			continue
		}

		// Extract extra_info (required)
		if extraInfoMatches := extraInfoPattern.FindStringSubmatch(obj); len(extraInfoMatches) > 1 {
			content.ExtraInfo = extraInfoMatches[1]
		} else {
			content.ExtraInfo = "" // Set empty if not found
		}

		// Extract type (required)
		if typeMatches := typePattern.FindStringSubmatch(obj); len(typeMatches) > 1 {
			content.Type = typeMatches[1]
			if content.Type != "movie" && content.Type != "series" {
				valid = false
				continue
			}
		} else {
			valid = false
			continue
		}

		// Extract rating (optional)
		if ratingMatches := ratingPattern.FindStringSubmatch(obj); len(ratingMatches) > 1 {
			if rating, err := strconv.ParseFloat(ratingMatches[1], 64); err == nil {
				content.Rating = &rating
			}
		}

		// Add to results if valid
		if valid {
			results = append(results, content)
		}
	}

	log.Printf("Manually extracted %d valid content items", len(results))
	return results
}

// truncateString truncates a string to a specified maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// buildContentExtractionPrompt creates the prompt for extracting content
func buildContentExtractionPrompt() string {
	return `You are a specialized JSON extraction tool. Extract movies and series from the provided text into a clean JSON array.

Each entry must follow this exact schema:
{
  "title": string,
  "year": number (for movies only, if available),
  "category": string ("Hollywood", "Foreign", "Anime", "TV Series"),
  "extra_info": string (e.g., "Download Hollywood Movie", "Episode 15–18 Added", "Complete"),
  "type": string ("movie" or "series"),
  "rating": number (optional, if available, on a scale of 1-10)
}

Critical rules:
1. Output ONLY the raw JSON array with no explanations, no markdown code blocks, and no backticks
2. Do not include Korean content
3. For movies, extract year as an integer if available
4. For series, ignore the year unless explicitly mentioned
5. Preserve episode/season information in extra_info
6. Ensure the output is valid parseable JSON with no additional text

Examples of correct format:
[{"title":"Movie 1","year":2023,"category":"Hollywood","extra_info":"Action","type":"movie"},{"title":"Series 1","category":"TV Series","extra_info":"Season 2","type":"series"}]

YOUR ENTIRE RESPONSE MUST BE A VALID JSON ARRAY ONLY. DO NOT INCLUDE ANY OTHER TEXT.
`
}
