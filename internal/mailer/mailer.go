package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"time"

	"github.com/go-mail/mail/v2"
)

// Embed the templates directory into the program.
// This directive makes the files available at runtime.
//
//go:embed templates/*
var templateFS embed.FS

// Mailer struct for SMTP connection and sender configuration.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New creates a new Mailer instance with SMTP credentials.
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send sends an email using the provided template and data.
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// Load the template from the embedded file system.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		log.Printf("Error parsing email template: %v", err)
		return err
	}

	// Render the subject, plain text body, and HTML body.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		log.Printf("Error executing subject template: %v", err)
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		log.Printf("Error executing plain body template: %v", err)
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		log.Printf("Error executing HTML body template: %v", err)
		return err
	}

	// Construct and send the email message.
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Attempt to send the email up to 3 times.
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			log.Printf("Email sent successfully to %s", recipient)
			return nil
		}
		log.Printf("Attempt %d to send email to %s failed: %v", i, recipient, err)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Failed to send email to %s after 3 attempts", recipient)
	return err
}
