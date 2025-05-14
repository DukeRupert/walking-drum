// internal/services/email_service.go
package services

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/dukerupert/walking-drum/internal/config"
	"github.com/rs/zerolog"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, templateName string, data map[string]interface{}) error
}

// emailService implements the EmailService interface
type emailService struct {
	config config.EmailConfig
	logger zerolog.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.EmailConfig, logger *zerolog.Logger) EmailService {
	return &emailService{
		config: cfg,
		logger: logger.With().Str("component", "email_service").Logger(),
	}
}

// SendEmail sends an email to the specified recipient
func (s *emailService) SendEmail(ctx context.Context, to, subject, templateName string, data map[string]interface{}) error {
	s.logger.Debug().
		Str("function", "emailService.SendEmail").
		Str("to", to).
		Str("subject", subject).
		Str("template", templateName).
		Msg("Sending email")

	// Load the email template
	tmpl, err := s.loadTemplate(templateName)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "emailService.SendEmail").
			Str("template", templateName).
			Msg("Failed to load email template")
		return fmt.Errorf("failed to load email template: %w", err)
	}

	// Render the template
	renderedHTML, err := s.renderTemplate(tmpl, data)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "emailService.SendEmail").
			Str("template", templateName).
			Msg("Failed to render email template")
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send the email based on the configured provider
	switch s.config.Provider {
	case "sendgrid":
		err = s.sendWithSendgrid(to, subject, renderedHTML)
	case "ses":
		err = s.sendWithSES(to, subject, renderedHTML)
	default:
		err = fmt.Errorf("unsupported email provider: %s", s.config.Provider)
	}

	if err != nil {
		s.logger.Error().
			Err(err).
			Str("function", "emailService.SendEmail").
			Str("to", to).
			Str("subject", subject).
			Str("provider", s.config.Provider).
			Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info().
		Str("function", "emailService.SendEmail").
		Str("to", to).
		Str("subject", subject).
		Str("template", templateName).
		Str("provider", s.config.Provider).
		Msg("Email sent successfully")

	return nil
}

// loadTemplate loads an email template from the file system
func (s *emailService) loadTemplate(templateName string) (*template.Template, error) {
	templatePath := filepath.Join(s.config.TemplateDir, templateName+".html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// renderTemplate renders an email template with the provided data
func (s *emailService) renderTemplate(tmpl *template.Template, data map[string]interface{}) (string, error) {
	// Add default data
	if data == nil {
		data = make(map[string]interface{})
	}

	// If needed, add common data here
	// data["baseUrl"] = s.config.BaseURL

	// Render the template
	buffer := new(strings.Builder)
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// sendWithSendgrid sends an email using the SendGrid API
func (s *emailService) sendWithSendgrid(to, subject, htmlContent string) error {
	// Implementation using SendGrid
	// This is a stub - you would implement the actual SendGrid API call here
	s.logger.Debug().
		Str("function", "emailService.sendWithSendgrid").
		Str("to", to).
		Str("subject", subject).
		Msg("Sending email with SendGrid")

	// Simulating successful sending
	return nil
}

// sendWithSES sends an email using AWS SES
func (s *emailService) sendWithSES(to, subject, htmlContent string) error {
	// Implementation using AWS SES
	// This is a stub - you would implement the actual AWS SES API call here
	s.logger.Debug().
		Str("function", "emailService.sendWithSES").
		Str("to", to).
		Str("subject", subject).
		Msg("Sending email with AWS SES")

	// Simulating successful sending
	return nil
}