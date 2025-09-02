package notifier

import (
	"bytes"
	"cine-pulse/storage"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	gomail "gopkg.in/mail.v2"
)

// EmailNotifier handles sending email notifications
type EmailNotifier struct {
	smtpHost       string
	smtpPort       int
	senderEmail    string
	senderPass     string
	recipientEmail string
	htmlTemplate   *template.Template
}

// EmailConfig contains configuration for email notifications
type EmailConfig struct {
	SMTPHost       string
	SMTPPort       int
	SenderEmail    string
	SenderPassword string
	RecipientEmail string
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config EmailConfig) (*EmailNotifier, error) {
	// Initialize HTML template for emails
	tmpl, err := template.New("email").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Cine Pulse - New Content Update</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 800px; margin: 0 auto; }
        h1 { color: #e50914; }
        h2 { color: #0071c5; margin-top: 30px; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
        th { background-color: #f4f4f4; text-align: left; padding: 10px; }
        td { padding: 10px; border-bottom: 1px solid #ddd; }
        .movie { background-color: #fff3e0; }
        .series { background-color: #e3f2fd; }
        .footer { font-size: 12px; color: #666; margin-top: 50px; text-align: center; }
        .count { font-weight: bold; color: #e50914; }
        .source { font-style: italic; color: #666; }
    </style>
</head>
<body>
    <h1>Cine Pulse - Content Update</h1>
    <p>The following content was scraped on {{.Date}} from {{.SourcesCount}} source(s).</p>
    
    <p>Total content scraped: <span class="count">{{.TotalCount}}</span></p>

    {{if .Movies}}
    <h2>Movies ({{len .Movies}})</h2>
    <table>
        <tr>
            <th>Title</th>
            <th>Year</th>
            <th>Category</th>
            <th>Extra Info</th>
            {{if .HasRatings}}<th>Rating</th>{{end}}
        </tr>
        {{range .Movies}}
        <tr class="movie">
            <td>{{.Title}}</td>
            <td>{{if .Year}}{{.Year}}{{else}}-{{end}}</td>
            <td>{{.Category}}</td>
            <td>{{.ExtraInfo}}</td>
            {{if $.HasRatings}}<td>{{if .Rating}}{{.Rating}}/10{{else}}-{{end}}</td>{{end}}
        </tr>
        {{end}}
    </table>
    {{end}}

    {{if .Series}}
    <h2>Series ({{len .Series}})</h2>
    <table>
        <tr>
            <th>Title</th>
            <th>Category</th>
            <th>Extra Info</th>
            {{if .HasRatings}}<th>Rating</th>{{end}}
        </tr>
        {{range .Series}}
        <tr class="series">
            <td>{{.Title}}</td>
            <td>{{.Category}}</td>
            <td>{{.ExtraInfo}}</td>
            {{if $.HasRatings}}<td>{{if .Rating}}{{.Rating}}/10{{else}}-{{end}}</td>{{end}}
        </tr>
        {{end}}
    </table>
    {{end}}

    <div class="source">
        <p>Source(s): {{.SourceURLs}}</p>
    </div>

    <div class="footer">
        <p>This is an automated email from Cine Pulse. Please do not reply.</p>
    </div>
</body>
</html>
`)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email template: %v", err)
	}

	return &EmailNotifier{
		smtpHost:       config.SMTPHost,
		smtpPort:       config.SMTPPort,
		senderEmail:    config.SenderEmail,
		senderPass:     config.SenderPassword,
		recipientEmail: config.RecipientEmail,
		htmlTemplate:   tmpl,
	}, nil
}

// GetEmailConfigFromEnv loads email configuration from environment variables
func GetEmailConfigFromEnv() EmailConfig {
	// Parse SMTP port with default value of 587 if not specified or invalid
	smtpPort := 587
	if portStr := os.Getenv("EMAIL_SMTP_PORT"); portStr != "" {
		if p, err := fmt.Sscanf(portStr, "%d", &smtpPort); err != nil || p != 1 {
			log.Printf("Invalid SMTP port '%s', using default 587", portStr)
			smtpPort = 587
		}
	}

	smtpHost := os.Getenv("EMAIL_SMTP_HOST")
	senderEmail := os.Getenv("EMAIL_SENDER")
	password := os.Getenv("EMAIL_PASSWORD")
	recipient := os.Getenv("EMAIL_RECIPIENT")

	// Log configuration (without showing full password)
	passwordDisplay := ""
	if len(password) > 0 {
		if len(password) > 8 {
			passwordDisplay = password[:4] + "..." + password[len(password)-4:]
		} else {
			passwordDisplay = "***"
		}
	}

	log.Printf("Email Configuration: Host=%s, Port=%d, Sender=%s, Token=%s, Recipient=%s",
		smtpHost, smtpPort, senderEmail, passwordDisplay, recipient)

	return EmailConfig{
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		SenderEmail:    senderEmail,
		SenderPassword: password,
		RecipientEmail: recipient,
	}
}

// NotifyContentUpdate sends an email with details about newly scraped content
func (n *EmailNotifier) NotifyContentUpdate(contents []storage.Content, sourceURLs []string) error {
	if len(contents) == 0 {
		log.Println("No content to notify about")
		return nil
	}

	if n.recipientEmail == "" {
		log.Println("No recipient email configured, skipping notification")
		return nil
	}

	// Debug information for troubleshooting
	log.Printf("Email configuration - SMTP Host: %s, Port: %d, Sender: %s, Auth Credentials Length: %d chars",
		n.smtpHost, n.smtpPort, n.senderEmail, len(n.senderPass))

	// Prepare data for template
	var movies []storage.Content
	var series []storage.Content
	hasRatings := false

	// Separate movies and series, and check for ratings
	for _, content := range contents {
		if content.Rating != nil {
			hasRatings = true
		}

		if content.Type == "movie" {
			movies = append(movies, content)
		} else if content.Type == "series" {
			series = append(series, content)
		}
	}

	// Prepare template data
	data := struct {
		Date         string
		TotalCount   int
		Movies       []storage.Content
		Series       []storage.Content
		HasRatings   bool
		SourcesCount int
		SourceURLs   string
	}{
		Date:         time.Now().Format("January 2, 2006 at 3:04 PM"),
		TotalCount:   len(contents),
		Movies:       movies,
		Series:       series,
		HasRatings:   hasRatings,
		SourcesCount: len(sourceURLs),
		SourceURLs:   strings.Join(sourceURLs, ", "),
	}

	// Render email template
	var emailBody bytes.Buffer
	if err := n.htmlTemplate.Execute(&emailBody, data); err != nil {
		return fmt.Errorf("failed to render email template: %v", err)
	}

	// Create a new message using gomail
	m := gomail.NewMessage()

	// Set email headers
	m.SetHeader("From", n.senderEmail)
	m.SetHeader("To", n.recipientEmail)
	m.SetHeader("Subject", fmt.Sprintf("Cine Pulse: %d New Content Items (%d Movies, %d Series)",
		len(contents), len(movies), len(series)))

	// Set both plain text and HTML versions
	plainText := fmt.Sprintf(
		"Cine Pulse Content Update\n\n"+
			"New content scraped on %s from %d source(s).\n"+
			"Total items: %d (%d movies, %d series)\n\n"+
			"Sources: %s\n\n"+
			"This is an automated email from Cine Pulse. Please do not reply.",
		data.Date, data.SourcesCount, data.TotalCount, len(movies), len(series), data.SourceURLs)

	m.SetBody("text/plain", plainText)
	m.AddAlternative("text/html", emailBody.String())

	// Setup dialer with Mailtrap SMTP credentials
	// For Mailtrap, username should be "api" and password should be your API token
	d := gomail.NewDialer(n.smtpHost, n.smtpPort, "api", n.senderPass)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Email notification sent to %s with %d content items",
		n.recipientEmail, len(contents))
	return nil
}
