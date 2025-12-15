package email

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/darkrockmountain/gomail"
	"github.com/darkrockmountain/gomail/providers/smtp"
	"github.com/flosch/pongo2/v5"
	appconfig "server/internal/config"
)

// EmailConfig holds configuration for email providers
type EmailConfig struct {
	Provider     string `mapstructure:"EMAIL_PROVIDER"`
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`

	// SES Configuration
	SESRegion        string `mapstructure:"SES_REGION"`
	SESAccessKeyID   string `mapstructure:"SES_ACCESS_KEY_ID"`
	SESSecretAccessKey string `mapstructure:"SES_SECRET_ACCESS_KEY"`

	// SendGrid Configuration
	SendGridAPIKey string `mapstructure:"SENDGRID_API_KEY"`

	// From configuration
	FromEmail string `mapstructure:"FROM_EMAIL"`
	FromName  string `mapstructure:"FROM_NAME"`

	// Template configuration
	TemplatePath string `mapstructure:"EMAIL_TEMPLATE_PATH"`
}

// EmailSender interface defines the contract for email sending
type EmailSender interface {
	SendEmail(ctx context.Context, email *EmailMessage) error
	SendEmailAsync(ctx context.Context, email *EmailMessage) <-chan error
}

// TemplateManager interface defines the contract for template management
type TemplateManager interface {
	RenderTemplate(templateName string, data interface{}) (string, error)
	RenderMJMLToHTML(mjmlTemplate string, data interface{}) (string, error)
	AddTemplate(name string, content string) error
}

// EmailMessage represents an email message
type EmailMessage struct {
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	HTML    string
	Text    string
	From    string
	ReplyTo string
}

// TemplateData represents the data structure for email templates
type TemplateData struct {
	Data         map[string]interface{}
	BaseURL      string
	TemplateName string
}

// EmailClient is the main email client implementation
type EmailClient struct {
	config         *EmailConfig
	templateMgr    TemplateManager
	emailSender    gomail.EmailSender
	templates      sync.Map // Thread-safe template cache
	mu            sync.RWMutex
}

// NewEmailClient creates a new email client
func NewEmailClient(cfg *appconfig.Config) (*EmailClient, error) {
	emailCfg := &EmailConfig{
		Provider:       "dummy", // Default to dummy provider
		TemplatePath:   "./templates/emails",
	}

	// Override with actual config if present
	if cfg != nil {
		emailCfg.Provider = cfg.EmailProvider
		emailCfg.SMTPHost = cfg.SMTPHost
		emailCfg.SMTPPort = cfg.SMTPPort
		emailCfg.SMTPUsername = cfg.SMTPUsername
		emailCfg.SMTPPassword = cfg.SMTPPassword
		emailCfg.FromEmail = cfg.FromEmail
		emailCfg.FromName = cfg.FromName
		emailCfg.TemplatePath = cfg.EmailTemplatePath

		// SES configuration
		emailCfg.SESRegion = cfg.SESRegion
		emailCfg.SESAccessKeyID = cfg.SESAccessKeyID
		emailCfg.SESSecretAccessKey = cfg.SESSecretAccessKey

		// SendGrid configuration
		emailCfg.SendGridAPIKey = cfg.SendGridAPIKey
	}

	client := &EmailClient{
		config:      emailCfg,
		templateMgr: NewPongoTemplateManager(emailCfg.TemplatePath),
	}

	// Initialize provider based on configuration
	if err := client.initializeProvider(); err != nil {
		return nil, fmt.Errorf("failed to initialize email provider: %w", err)
	}

	return client, nil
}

// initializeProvider sets up the email provider based on configuration
func (ec *EmailClient) initializeProvider() error {
	switch ec.config.Provider {
	case "smtp":
		return ec.initializeSMTP()
	case "ses":
		return ec.initializeSES()
	case "sendgrid":
		return ec.initializeSendGrid()
	case "dummy":
		// Dummy provider for development/testing
		return nil
	default:
		return fmt.Errorf("unsupported email provider: %s", ec.config.Provider)
	}
}

// initializeSMTP configures SMTP sender
func (ec *EmailClient) initializeSMTP() error {
	if ec.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required for SMTP provider")
	}

	port := 587 // Default SMTP port
	if ec.config.SMTPPort > 0 {
		port = ec.config.SMTPPort
	}

	// Create SMTP sender using gomail SMTP provider
	smtpSender, err := smtp.NewSmtpEmailSender(ec.config.SMTPHost, port, ec.config.SMTPUsername, ec.config.SMTPPassword, smtp.AUTH_PLAIN)
	if err != nil {
		return fmt.Errorf("failed to create SMTP sender: %w", err)
	}
	ec.emailSender = smtpSender
	return nil
}

// initializeSES configures AWS SES client
func (ec *EmailClient) initializeSES() error {
	// TODO: Implement SES client initialization
	// This would require AWS SDK v2 for SES
	return fmt.Errorf("SES provider not yet implemented")
}

// initializeSendGrid configures SendGrid client
func (ec *EmailClient) initializeSendGrid() error {
	// TODO: Implement SendGrid client initialization
	// This would require SendGrid Go SDK
	return fmt.Errorf("SendGrid provider not yet implemented")
}

// SendEmail sends an email synchronously
func (ec *EmailClient) SendEmail(ctx context.Context, email *EmailMessage) error {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	switch ec.config.Provider {
	case "smtp":
		return ec.sendViaSMTP(ctx, email)
	case "ses":
		return ec.sendViaSES(ctx, email)
	case "sendgrid":
		return ec.sendViaSendGrid(ctx, email)
	case "dummy":
		return ec.sendViaDummy(ctx, email)
	default:
		return fmt.Errorf("unsupported email provider: %s", ec.config.Provider)
	}
}

// SendEmailAsync sends an email asynchronously
func (ec *EmailClient) SendEmailAsync(ctx context.Context, email *EmailMessage) <-chan error {
	errorChan := make(chan error, 1)

	go func() {
		defer close(errorChan)
		errorChan <- ec.SendEmail(ctx, email)
	}()

	return errorChan
}

// SendTemplateEmail sends an email using a template
func (ec *EmailClient) SendTemplateEmail(ctx context.Context, templateName string, data interface{}, email *EmailMessage) error {
	htmlContent, err := ec.templateMgr.RenderTemplate(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Try to render text template if it exists
	textTemplateName := fmt.Sprintf("%s.txt", templateName)
	textContent, err := ec.templateMgr.RenderTemplate(textTemplateName, data)
	if err != nil {
		// If text template doesn't exist, create a simple text version
		textContent = ec.extractTextFromHTML(htmlContent)
	}

	email.HTML = htmlContent
	email.Text = textContent

	return ec.SendEmail(ctx, email)
}

// SendTemplateEmailAsync sends a template email asynchronously
func (ec *EmailClient) SendTemplateEmailAsync(ctx context.Context, templateName string, data interface{}, email *EmailMessage) <-chan error {
	errorChan := make(chan error, 1)

	go func() {
		defer close(errorChan)
		errorChan <- ec.SendTemplateEmail(ctx, templateName, data, email)
	}()

	return errorChan
}

// Helper methods for different providers
func (ec *EmailClient) sendViaSMTP(ctx context.Context, email *EmailMessage) error {
	// Determine from address
	fromAddress := email.From
	if fromAddress == "" && ec.config.FromEmail != "" {
		if ec.config.FromName != "" {
			fromAddress = fmt.Sprintf("%s <%s>", ec.config.FromName, ec.config.FromEmail)
		} else {
			fromAddress = ec.config.FromEmail
		}
	}

	// Create gomail message
	msg := gomail.NewFullEmailMessage(
		fromAddress,
		email.To,
		email.Subject,
		email.Cc,
		email.Bcc,
		email.ReplyTo,
		email.Text, // textBody
		email.HTML, // htmlBody
		nil, // attachments
	)

	// Send the email using the email sender
	return ec.emailSender.SendEmail(msg)
}

func (ec *EmailClient) sendViaSES(ctx context.Context, email *EmailMessage) error {
	// TODO: Implement SES sending
	return fmt.Errorf("SES provider not yet implemented")
}

func (ec *EmailClient) sendViaSendGrid(ctx context.Context, email *EmailMessage) error {
	// TODO: Implement SendGrid sending
	return fmt.Errorf("SendGrid provider not yet implemented")
}

func (ec *EmailClient) sendViaDummy(ctx context.Context, email *EmailMessage) error {
	// Dummy implementation for development/testing
	fmt.Printf("Dummy Email Sender:\n")
	fmt.Printf("To: %v\n", email.To)
	fmt.Printf("Subject: %s\n", email.Subject)
	fmt.Printf("HTML: %s\n", email.HTML)
	fmt.Printf("Text: %s\n", email.Text)
	return nil
}

// extractTextFromHTML creates a simple text version from HTML
func (ec *EmailClient) extractTextFromHTML(html string) string {
	// Simple HTML to text conversion
	// In a real implementation, you might use a proper HTML to text library
	text := html

	// Remove common HTML tags
	replacements := map[string]string{
		"<br>":    "\n",
		"</br>":   "\n",
		"<p>":     "",
		"</p>":    "\n\n",
		"<div>":   "",
		"</div>":  "\n",
		"<h1>":    "\n# ",
		"</h1>":   "\n\n",
		"<h2>":    "\n## ",
		"</h2>":   "\n\n",
		"<h3>":    "\n### ",
		"</h3>":   "\n\n",
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	return text
}

// PongoTemplateManager implements TemplateManager using pongo2
type PongoTemplateManager struct {
	templatePath string
	templates    sync.Map
	loader       pongo2.TemplateLoader
}

// NewPongoTemplateManager creates a new pongo2-based template manager
func NewPongoTemplateManager(templatePath string) *PongoTemplateManager {
	// Create a template loader that can resolve relative paths
	loader, _ := pongo2.NewLocalFileSystemLoader(templatePath)
	return &PongoTemplateManager{
		templatePath: templatePath,
		loader:       loader,
	}
}

// RenderTemplate renders a template using pongo2
func (ptm *PongoTemplateManager) RenderTemplate(templateName string, data interface{}) (string, error) {
	// Convert data to pongo2.Context
	ctx, ok := data.(map[string]interface{})
	if !ok {
		// Try to convert to map[string]interface{}
		ctx = map[string]interface{}{"data": data}
	}

	// Try to get from cache first
	if cached, ok := ptm.templates.Load(templateName); ok {
		if tmpl, ok := cached.(*pongo2.Template); ok {
			return tmpl.Execute(ctx)
		}
	}

	// Load template from filesystem
	tmpl, err := pongo2.FromFile(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", templateName, err)
	}

	// Cache the template
	ptm.templates.Store(templateName, tmpl)

	return tmpl.Execute(ctx)
}

// RenderMJMLToHTML renders MJML template to HTML
func (ptm *PongoTemplateManager) RenderMJMLToHTML(mjmlTemplate string, data interface{}) (string, error) {
	// Convert data to pongo2.Context
	ctx, ok := data.(map[string]interface{})
	if !ok {
		// Try to convert to map[string]interface{}
		ctx = map[string]interface{}{"data": data}
	}

	// For now, just render the MJML as pongo2 template
	// In a full implementation, you would use an MJML to HTML converter
	tmpl, err := pongo2.FromString(mjmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse MJML template: %w", err)
	}

	return tmpl.Execute(ctx)
}

// AddTemplate adds a template to the manager
func (ptm *PongoTemplateManager) AddTemplate(name string, content string) error {
	tmpl, err := pongo2.FromString(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	ptm.templates.Store(name, tmpl)
	return nil
}

// NewEmailClientProvider creates an email client using configuration for FX
// This follows the same pattern as NewS3ClientProvider for dependency injection
func NewEmailClientProvider(cfg *appconfig.Config) (EmailSender, error) {
	// Return nil client gracefully if email is not configured
	// This allows the service to work in development without email setup
	if cfg.EmailProvider == "" {
		return nil, nil
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
