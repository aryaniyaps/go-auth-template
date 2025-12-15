package email

import (
	"context"

	appconfig "server/internal/config"
)

// EmailTemplateData provides common data structure for email templates
type EmailTemplateData struct {
	// Common application data
	AppName     string
	AppURL      string
	SupportEmail string

	// Template specific data
	Data map[string]interface{}
}

// NewEmailTemplateData creates a new template data structure with common fields
func NewEmailTemplateData(cfg *appconfig.Config) *EmailTemplateData {
	data := &EmailTemplateData{
		AppName:      "HospitalJobs", // Default app name, should be configurable
		AppURL:       "https://hospitaljobsin.com", // Default URL, should be configurable
		SupportEmail: "support@hospitaljobsin.com", // Default support email
		Data:         make(map[string]interface{}),
	}

	// Override with config if available
	if cfg != nil {
		// TODO: Add these fields to config when available
		// data.AppName = cfg.AppName
		// data.AppURL = cfg.AppURL
		// data.SupportEmail = cfg.SupportEmail
	}

	return data
}

// SetField sets a field in the template data
func (etd *EmailTemplateData) SetField(key string, value interface{}) {
	etd.Data[key] = value
}

// GetField gets a field from the template data
func (etd *EmailTemplateData) GetField(key string) interface{} {
	return etd.Data[key]
}

// ToMap converts the template data to a map for Pongo2 rendering
func (etd *EmailTemplateData) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"app_name":      etd.AppName,
		"app_url":       etd.AppURL,
		"support_email": etd.SupportEmail,
	}

	// Merge with custom data
	for k, v := range etd.Data {
		result[k] = v
	}

	return result
}

// EmailVerificationData creates template data for email verification
func EmailVerificationData(cfg *appconfig.Config, email, token, userAgent string) map[string]interface{} {
	data := NewEmailTemplateData(cfg)

	data.SetField("email", email)
	data.SetField("verification_token", token)
	data.SetField("token_expires_in", "15 minutes") // TODO: Make configurable
	data.SetField("user_agent", userAgent)

	return data.ToMap()
}

// PasswordResetData creates template data for password reset
func PasswordResetData(cfg *appconfig.Config, resetLink, userAgent string, isInitial bool) map[string]interface{} {
	data := NewEmailTemplateData(cfg)

	data.SetField("reset_link", resetLink)
	data.SetField("link_expires_in", "1 hour") // TODO: Make configurable
	data.SetField("user_agent", userAgent)
	data.SetField("is_initial", isInitial)

	return data.ToMap()
}

// RenderSubject renders an email subject template
func (ec *EmailClient) RenderSubject(templateName string, data map[string]interface{}) (string, error) {
	// For simple templates like subjects, use direct string rendering
	// This avoids path resolution issues with complex template inheritance
	return ec.templateMgr.RenderTemplate(templateName, data)
}

// RenderEmail renders both HTML and text versions of an email
func (ec *EmailClient) RenderEmail(templatePath string, data map[string]interface{}) (htmlContent, textContent string, err error) {
	// Render HTML version
	htmlTemplate := templatePath + "/body.mjml"
	htmlContent, err = ec.templateMgr.RenderMJMLToHTML(htmlTemplate, data)
	if err != nil {
		return "", "", err
	}

	// Render text version
	textTemplate := templatePath + "/body.txt"
	textContent, err = ec.templateMgr.RenderTemplate(textTemplate, data)
	if err != nil {
		// If text template fails, create a basic text version
		textContent = ec.extractTextFromHTML(htmlContent)
	}

	return htmlContent, textContent, nil
}

// SendEmailTemplate sends an email using template files for subject, HTML, and text
func (ec *EmailClient) SendEmailTemplate(ctx context.Context, templatePath string, data map[string]interface{}, to []string) error {
	// Render subject
	subject, err := ec.RenderSubject(templatePath+"/subject.txt", data)
	if err != nil {
		return err
	}

	// Render email content
	htmlContent, textContent, err := ec.RenderEmail(templatePath, data)
	if err != nil {
		return err
	}

	// Create and send message
	message := &EmailMessage{
		To:      to,
		Subject: subject,
		HTML:    htmlContent,
		Text:    textContent,
	}

	return ec.SendEmail(ctx, message)
}

// SendEmailVerification sends an email verification email
func (ec *EmailClient) SendEmailVerification(ctx context.Context, cfg *appconfig.Config, email, token, userAgent string) error {
	data := EmailVerificationData(cfg, email, token, userAgent)
	return ec.SendEmailTemplate(ctx, "emails/email-verification", data, []string{email})
}

// SendPasswordReset sends a password reset email
func (ec *EmailClient) SendPasswordReset(ctx context.Context, cfg *appconfig.Config, resetLink, userAgent string, isInitial bool, toEmail string) error {
	data := PasswordResetData(cfg, resetLink, userAgent, isInitial)
	return ec.SendEmailTemplate(ctx, "emails/password-reset", data, []string{toEmail})
}