package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"mime"
	"net"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// Config holds email configuration
type Config struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	FromName string
	UseTLS   bool
}

// Service handles email operations
type Service struct {
	config Config
}

// NewService creates a new email service
func NewService(config Config) *Service {
	return &Service{config: config}
}

// NewServiceFromSettings creates email service from app settings
func NewServiceFromSettings(settings models.AppSettingModelInterface) (*Service, error) {
	enabled, err := settings.GetBool("email_enabled")
	if err != nil {
		enabled = false
	}

	host, err := settings.GetString("smtp_host")
	if err != nil {
		host = "smtp.gmail.com"
	}

	port, err := settings.GetInt("smtp_port")
	if err != nil {
		port = 587
	}

	username, err := settings.GetString("smtp_username")
	if err != nil {
		username = ""
	}

	password, err := settings.GetDecryptedSMTPPassword()
	if err != nil {
		password = ""
	}

	fromName, err := settings.GetString("smtp_from_name")
	if err != nil {
		fromName = "FreelanceTracker"
	}

	useTLS, err := settings.GetBool("smtp_use_tls")
	if err != nil {
		useTLS = true
	}

	config := Config{
		Enabled:  enabled,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		FromName: fromName,
		UseTLS:   useTLS,
	}

	return NewService(config), nil
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Data     []byte
	MimeType string
}

// Email represents an email message
type Email struct {
	To          []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []Attachment
}

// Send sends an email using SMTP
func (s *Service) Send(email Email) error {
	if !s.config.Enabled {
		return fmt.Errorf("email service is disabled")
	}

	if s.config.Username == "" || s.config.Password == "" {
		return fmt.Errorf("email credentials not configured")
	}

	// Create message
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.Username)
	to := strings.Join(email.To, ", ")

	var msg string
	if len(email.Attachments) == 0 {
		// Simple message without attachments
		var contentType string
		if email.IsHTML {
			contentType = "text/html; charset=UTF-8"
		} else {
			contentType = "text/plain; charset=UTF-8"
		}

		msg = fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: %s\r\n"+
			"\r\n"+
			"%s",
			from, to, email.Subject, contentType, email.Body)
	} else {
		// Multipart message with attachments
		boundary := s.generateBoundary()
		
		// Email headers
		msg = fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/mixed; boundary=\"%s\"\r\n"+
			"\r\n",
			from, to, email.Subject, boundary)

		// Body part
		var bodyContentType string
		if email.IsHTML {
			bodyContentType = "text/html; charset=UTF-8"
		} else {
			bodyContentType = "text/plain; charset=UTF-8"
		}

		msg += fmt.Sprintf("--%s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Transfer-Encoding: 7bit\r\n"+
			"\r\n"+
			"%s\r\n\r\n",
			boundary, bodyContentType, email.Body)

		// Attachment parts
		for _, attachment := range email.Attachments {
			msg += s.createAttachmentPart(boundary, attachment)
		}

		// End boundary
		msg += fmt.Sprintf("--%s--\r\n", boundary)
	}

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	var auth smtp.Auth
	auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, s.config.Username, email.To, []byte(msg))
	}

	return smtp.SendMail(addr, auth, s.config.Username, email.To, []byte(msg))
}

// sendWithTLS sends email with STARTTLS encryption (for Gmail port 587)
func (s *Service) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Create plain connection first
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Start TLS if supported (STARTTLS)
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// TestConnection tests the SMTP connection
func (s *Service) TestConnection() error {
	if !s.config.Enabled {
		return fmt.Errorf("email service is disabled")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	if s.config.UseTLS {
		// Create plain connection first (for STARTTLS)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		// Create SMTP client
		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		// Start TLS if supported (STARTTLS)
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: s.config.Host,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("failed to start TLS: %w", err)
			}
		}

		if s.config.Username != "" && s.config.Password != "" {
			auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
		}
	} else {
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer client.Quit()
	}

	return nil
}

// generateBoundary generates a unique boundary string for multipart messages
func (s *Service) generateBoundary() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("----=_Part_%d_%d", time.Now().Unix(), rand.Intn(1000000))
}

// createAttachmentPart creates a MIME part for an attachment
func (s *Service) createAttachmentPart(boundary string, attachment Attachment) string {
	// Determine MIME type if not provided
	mimeType := attachment.MimeType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(attachment.Filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	// Encode attachment data to base64
	encodedData := base64.StdEncoding.EncodeToString(attachment.Data)
	
	// Split base64 data into lines (RFC 2045 recommends max 76 characters per line)
	var lines []string
	for i := 0; i < len(encodedData); i += 76 {
		end := i + 76
		if end > len(encodedData) {
			end = len(encodedData)
		}
		lines = append(lines, encodedData[i:end])
	}
	
	return fmt.Sprintf("--%s\r\n"+
		"Content-Type: %s; name=\"%s\"\r\n"+
		"Content-Transfer-Encoding: base64\r\n"+
		"Content-Disposition: attachment; filename=\"%s\"\r\n"+
		"\r\n"+
		"%s\r\n\r\n",
		boundary, mimeType, attachment.Filename, attachment.Filename, strings.Join(lines, "\r\n"))
}

// SendPaymentReceivedEmail sends a payment received notification email to both client and freelancer
func (s *Service) SendPaymentReceivedEmail(clientEmail, clientName, freelancerEmail, freelancerName, projectName string) error {
	if !s.config.Enabled {
		return fmt.Errorf("email service is disabled")
	}

	// Extract first name from client name (take first word)
	clientFirstName := clientName
	if fields := strings.Fields(clientName); len(fields) > 0 {
		clientFirstName = fields[0]
	}

	// Extract first name from freelancer name (take first word)
	freelancerFirstName := freelancerName
	if fields := strings.Fields(freelancerName); len(fields) > 0 {
		freelancerFirstName = fields[0]
	}

	subject := "Payment for Academic Editing Received"
	body := fmt.Sprintf("Dear %s,\n\nWe have received payment for your project: %s\n\nThank you for your business!\n\nBest regards,\n%s",
		clientFirstName, projectName, freelancerFirstName)

	// Send to both client and freelancer
	recipients := []string{clientEmail}
	if freelancerEmail != "" && freelancerEmail != clientEmail {
		recipients = append(recipients, freelancerEmail)
	}

	email := Email{
		To:      recipients,
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}

	return s.Send(email)
}