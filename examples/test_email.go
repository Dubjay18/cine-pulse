package examples

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/mail.v2"
)

// TestMailtrapEmail sends a test email via Mailtrap to verify configuration
func TestMailtrapEmail() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Log environment variables (with masked password)
	smtpHost := os.Getenv("EMAIL_SMTP_HOST")
	smtpPortStr := os.Getenv("EMAIL_SMTP_PORT")
	senderEmail := os.Getenv("EMAIL_SENDER")
	password := os.Getenv("EMAIL_PASSWORD")
	recipient := os.Getenv("EMAIL_RECIPIENT")

	// Mask password for logging
	passwordDisplay := ""
	if len(password) > 0 {
		if len(password) > 8 {
			passwordDisplay = password[:4] + "..." + password[len(password)-4:]
		} else {
			passwordDisplay = "***"
		}
	}

	log.Printf("Email Configuration: Host=%s, Port=%s, Sender=%s, Token=%s, Recipient=%s",
		smtpHost, smtpPortStr, senderEmail, passwordDisplay, recipient)

	// Create a new message
	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", "Test Email from Cine-Pulse")
	m.SetBody("text/html", "<h1>Test Email</h1><p>This is a test email from Cine-Pulse to verify Mailtrap configuration.</p>")

	// Create dialer
	port := 587
	if smtpPortStr != "" {
		if _, err := fmt.Sscanf(smtpPortStr, "%d", &port); err != nil {
			log.Printf("Invalid port number: %s, using default 587", smtpPortStr)
		}
	}

	log.Printf("Creating dialer with: Host=%s, Port=%d, Username=%s", smtpHost, port, "api")
	d := mail.NewDialer(smtpHost, port, "api", password)
	d.SSL = false

	// Send email
	log.Println("Attempting to send test email...")
	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Println("Email sent successfully!")
}
