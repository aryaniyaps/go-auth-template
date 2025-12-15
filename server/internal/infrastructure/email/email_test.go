package email

import (
	"context"
	"strings"
	"testing"

	"github.com/flosch/pongo2/v5"
	appconfig "server/internal/config"
)

// MockEmailSender implements EmailSender interface for testing
type MockEmailSender struct {
	sentEmails []*EmailMessage
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		sentEmails: make([]*EmailMessage, 0),
	}
}

func (m *MockEmailSender) SendEmail(ctx context.Context, email *EmailMessage) error {
	m.sentEmails = append(m.sentEmails, email)
	return nil
}

func (m *MockEmailSender) SendEmailAsync(ctx context.Context, email *EmailMessage) <-chan error {
	errorChan := make(chan error, 1)
	go func() {
		defer close(errorChan)
		m.sentEmails = append(m.sentEmails, email)
		errorChan <- nil
	}()
	return errorChan
}

func (m *MockEmailSender) GetSentEmails() []*EmailMessage {
	return m.sentEmails
}

func TestNewEmailClient(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}
}

func TestNewEmailClientProvider(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "",
	}

	sender, err := NewEmailClientProvider(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sender != nil {
		t.Fatal("Expected sender to be nil when provider is empty")
	}

	cfg.EmailProvider = "dummy"
	sender, err = NewEmailClientProvider(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sender == nil {
		t.Fatal("Expected sender to be non-nil when provider is set")
	}
}

func TestEmailMessage(t *testing.T) {
	msg := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Test Subject",
		Text:    "Test text content",
		HTML:    "<p>Test HTML content</p>",
	}

	if len(msg.To) != 1 || msg.To[0] != "test@example.com" {
		t.Error("To field not set correctly")
	}

	if msg.Subject != "Test Subject" {
		t.Error("Subject field not set correctly")
	}

	if msg.Text != "Test text content" {
		t.Error("Text field not set correctly")
	}

	if msg.HTML != "<p>Test HTML content</p>" {
		t.Error("HTML field not set correctly")
	}
}

func TestPongoTemplateManager(t *testing.T) {
	templateMgr := NewPongoTemplateManager("./templates")

	templateContent := `<h1>Hello {{name}}!</h1>`
	err := templateMgr.AddTemplate("test.html", templateContent)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	data := map[string]interface{}{
		"name": "World",
	}

	rendered, err := templateMgr.RenderTemplate("test.html", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "<h1>Hello World!</h1>"
	if rendered != expected {
		t.Errorf("Expected %q, got %q", expected, rendered)
	}

	mjmlContent := `<mjml><mj-body><mj-text>Hello {{name}}!</mj-text></mj-body></mjml>`
	renderedMJML, err := templateMgr.RenderMJMLToHTML(mjmlContent, data)
	if err != nil {
		t.Fatalf("Failed to render MJML template: %v", err)
	}

	if renderedMJML == "" {
		t.Error("Expected non-empty MJML output")
	}
}

func TestEmailClientSendEmail(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	ctx := context.Background()
	msg := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Test Email",
		Text:    "Test message",
		HTML:    "<p>Test message</p>",
	}

	err = client.SendEmail(ctx, msg)
	if err != nil {
		t.Errorf("Expected no error with dummy provider, got %v", err)
	}
}

func TestEmailClientSendTemplateEmail(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	templateContent := `<h1>Welcome {{UserName}}!</h1><p>{{Message}}</p>`
	err = client.templateMgr.AddTemplate("welcome.html", templateContent)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	ctx := context.Background()
	data := map[string]interface{}{
		"UserName": "Test User",
		"Message":  "Welcome to our service!",
	}

	msg := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Welcome Email",
	}

	err = client.SendTemplateEmail(ctx, "welcome.html", data, msg)
	if err != nil {
		t.Errorf("Expected no error with dummy provider, got %v", err)
	}
}

func TestEmailClientSendEmailAsync(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	ctx := context.Background()
	msg := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Async Test Email",
		Text:    "Test message",
	}

	errorChan := client.SendEmailAsync(ctx, msg)
	err = <-errorChan
	if err != nil {
		t.Errorf("Expected no error with dummy provider, got %v", err)
	}
}

func TestExtractTextFromHTML(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	html := `<h1>Title</h1><p>Paragraph content</p><br><div>Div content</div>`
	text := client.extractTextFromHTML(html)

	if text == "" {
		t.Error("Expected non-empty text output")
	}

	if len(text) > len(html)*2 {
		t.Error("Text output seems unusually long compared to HTML input")
	}
}

func TestSimpleTemplateRendering(t *testing.T) {
	subjectTemplate := "{{ app_name }} Email Verification Request"
	tmpl, err := pongo2.FromString(subjectTemplate)
	if err != nil {
		t.Fatalf("Failed to create template from string: %v", err)
	}

	data := map[string]interface{}{
		"app_name": "TestApp",
	}

	result, err := tmpl.Execute(data)
	if err != nil {
		t.Errorf("Failed to execute template: %v", err)
	}

	expected := "TestApp Email Verification Request"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	passwordResetTemplate := `
{% if is_initial %}
Set your {{ app_name }} password
{% else %}
{{ app_name }} Password Reset Request
{% endif %}`

	tmpl2, err := pongo2.FromString(passwordResetTemplate)
	if err != nil {
		t.Fatalf("Failed to create password reset template: %v", err)
	}

	data2 := map[string]interface{}{
		"app_name":   "TestApp",
		"is_initial": true,
	}

	result2, err := tmpl2.Execute(data2)
	if err != nil {
		t.Errorf("Failed to execute password reset template: %v", err)
	}

	expected2 := "Set your TestApp password"
	if strings.TrimSpace(result2) != expected2 {
		t.Errorf("Expected '%s', got '%s'", expected2, strings.TrimSpace(result2))
	}

	data3 := map[string]interface{}{
		"app_name":   "TestApp",
		"is_initial": false,
	}

	result3, err := tmpl2.Execute(data3)
	if err != nil {
		t.Errorf("Failed to execute password reset template (reset): %v", err)
	}

	expected3 := "TestApp Password Reset Request"
	if strings.TrimSpace(result3) != expected3 {
		t.Errorf("Expected '%s', got '%s'", expected3, strings.TrimSpace(result3))
	}
}

func TestTemplateDataHelpers(t *testing.T) {
	cfg := &appconfig.Config{}

	data := EmailVerificationData(cfg, "test@example.com", "123456", "Mozilla/5.0")

	if data["app_name"] != "HospitalJobs" {
		t.Error("App name not set correctly in email verification data")
	}
	if data["email"] != "test@example.com" {
		t.Error("Email not set correctly in email verification data")
	}
	if data["verification_token"] != "123456" {
		t.Error("Verification token not set correctly")
	}

	data2 := PasswordResetData(cfg, "https://example.com/reset", "Mozilla/5.0", true)

	if data2["is_initial"] != true {
		t.Error("Is initial flag not set correctly in password reset data")
	}
	if data2["reset_link"] != "https://example.com/reset" {
		t.Error("Reset link not set correctly")
	}
	if data2["link_expires_in"] != "1 hour" {
		t.Error("Link expiration not set correctly")
	}
}

func TestEmailClientWithDummyProvider(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider:     "dummy",
		EmailTemplatePath: "./templates/emails",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	templateData := EmailVerificationData(cfg, "test@example.com", "123456", "Mozilla/5.0")

	subject := templateData["app_name"].(string) + " Email Verification Request"

	htmlContent := `<h1>Email Verification</h1><p>Your verification code: ` +
		templateData["verification_token"].(string) + `</p>`

	textContent := "Use this verification code: " + templateData["verification_token"].(string)

	message := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: subject,
		HTML:    htmlContent,
		Text:    textContent,
	}

	err = client.SendEmail(context.TODO(), message)
	if err != nil {
		t.Errorf("Failed to send email with dummy provider: %v", err)
	}
}

func TestMJMLTemplateRendering(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider:     "dummy",
		EmailTemplatePath: "./templates/emails",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	mjmlTemplate := `
<mjml>
  <mj-head>
    <mj-title>{{ app_name }} Verification</mj-title>
  </mj-head>
  <mj-body>
    <mj-text>Hello {{ email }}!</mj-text>
    <mj-text>Your code: {{ verification_token }}</mj-text>
  </mj-body>
</mjml>`

	data := map[string]interface{}{
		"app_name":          "TestApp",
		"email":             "test@example.com",
		"verification_token": "123456",
	}

	result, err := client.templateMgr.RenderMJMLToHTML(mjmlTemplate, data)
	if err != nil {
		t.Errorf("Failed to render MJML template: %v", err)
	}

	if result == "" {
		t.Error("MJML template rendering returned empty result")
	}

	if len(result) < 50 {
		t.Error("MJML template result seems too short, template variables may not have been substituted")
	}
}

func TestEmailVerificationTemplate(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	// Test email verification data generation
	data := EmailVerificationData(cfg, "test@example.com", "123456", "Mozilla/5.0")

	// Verify all required fields are present
	if data["app_name"] != "HospitalJobs" {
		t.Error("App name not set correctly")
	}
	if data["email"] != "test@example.com" {
		t.Error("Email not set correctly")
	}
	if data["verification_token"] != "123456" {
		t.Error("Verification token not set correctly")
	}
	if data["user_agent"] != "Mozilla/5.0" {
		t.Error("User agent not set correctly")
	}
}

func TestPasswordResetTemplate(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	// Test password reset data generation for initial setup
	data1 := PasswordResetData(cfg, "https://example.com/reset", "Mozilla/5.0", true)

	if data1["is_initial"] != true {
		t.Error("Is initial flag not set correctly for initial setup")
	}
	if data1["reset_link"] != "https://example.com/reset" {
		t.Error("Reset link not set correctly")
	}

	// Test password reset data generation for reset
	data2 := PasswordResetData(cfg, "https://example.com/reset", "Mozilla/5.0", false)

	if data2["is_initial"] != false {
		t.Error("Is initial flag not set correctly for password reset")
	}
	if data2["reset_link"] != "https://example.com/reset" {
		t.Error("Reset link not set correctly")
	}
}

func TestEmailTemplateData(t *testing.T) {
	cfg := &appconfig.Config{}

	data := NewEmailTemplateData(cfg)

	data.SetField("test_key", "test_value")
	if data.GetField("test_key") != "test_value" {
		t.Error("Failed to set/get field correctly")
	}

	dataMap := data.ToMap()
	if dataMap["app_name"] != "HospitalJobs" {
		t.Error("Default app name not set correctly")
	}
	if dataMap["test_key"] != "test_value" {
		t.Error("Custom field not included in ToMap output")
	}
}

func TestEmailVerificationData(t *testing.T) {
	cfg := &appconfig.Config{}
	email := "test@example.com"
	token := "123456"
	userAgent := "Mozilla/5.0 (Test Browser)"

	data := EmailVerificationData(cfg, email, token, userAgent)

	if data["email"] != email {
		t.Error("Email field not set correctly")
	}
	if data["verification_token"] != token {
		t.Error("Verification token field not set correctly")
	}
	if data["user_agent"] != userAgent {
		t.Error("User agent field not set correctly")
	}
	if data["token_expires_in"] != "15 minutes" {
		t.Error("Token expiration field not set correctly")
	}
}

func TestPasswordResetData(t *testing.T) {
	cfg := &appconfig.Config{}
	resetLink := "https://example.com/reset?token=abc123"
	userAgent := "Mozilla/5.0 (Test Browser)"
	isInitial := true

	data := PasswordResetData(cfg, resetLink, userAgent, isInitial)

	if data["reset_link"] != resetLink {
		t.Error("Reset link field not set correctly")
	}
	if data["user_agent"] != userAgent {
		t.Error("User agent field not set correctly")
	}
	if data["is_initial"] != isInitial {
		t.Error("Is initial field not set correctly")
	}
	if data["link_expires_in"] != "1 hour" {
		t.Error("Link expiration field not set correctly")
	}

	data = PasswordResetData(cfg, resetLink, userAgent, false)
	if data["is_initial"] != false {
		t.Error("Is initial field not set correctly for false value")
	}
}

func TestRenderSubject(t *testing.T) {
	// Test that template rendering works with template manager
	templateMgr := NewPongoTemplateManager("./templates")
	testTemplate := "Hello {{name}}!"
	err := templateMgr.AddTemplate("test_subject.txt", testTemplate)
	if err != nil {
		t.Fatalf("Failed to add test template: %v", err)
	}

	data := map[string]interface{}{
		"name": "World",
	}

	// Test template rendering through template manager
	rendered, err := templateMgr.RenderTemplate("test_subject.txt", data)
	if err != nil {
		t.Errorf("Failed to render test template: %v", err)
	}

	expected := "Hello World!"
	if rendered != expected {
		t.Errorf("Expected '%s', got '%s'", expected, rendered)
	}
}

func TestRenderEmail(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	// Test HTML to text extraction
	html := "<h1>Hello</h1><p>World</p>"
	text := client.extractTextFromHTML(html)

	if text == "" {
		t.Error("Text extraction should not return empty string")
	}
}

func TestSendEmailTemplate(t *testing.T) {
	cfg := &appconfig.Config{
		EmailProvider: "dummy",
	}

	client, err := NewEmailClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create email client: %v", err)
	}

	ctx := context.Background()

	// Test creating a template message and sending it
	templateContent := "<h1>Hello {{name}}!</h1>"
	err = client.templateMgr.AddTemplate("test.html", templateContent)
	if err != nil {
		t.Fatalf("Failed to add template: %v", err)
	}

	data := map[string]interface{}{
		"name": "Test User",
	}

	message := &EmailMessage{
		To:      []string{"test@example.com"},
		Subject: "Template Test",
	}

	err = client.SendTemplateEmail(ctx, "test.html", data, message)
	if err != nil {
		t.Errorf("Failed to send template email: %v", err)
	}
}

func TestTemplateHelpers(t *testing.T) {
	cfg := &appconfig.Config{}
	templateData := NewEmailTemplateData(cfg)

	if templateData.AppName != "HospitalJobs" {
		t.Errorf("Expected app name 'HospitalJobs', got '%s'", templateData.AppName)
	}

	if templateData.AppURL != "https://hospitaljobsin.com" {
		t.Errorf("Expected app URL 'https://hospitaljobsin.com', got '%s'", templateData.AppURL)
	}

	if templateData.SupportEmail != "support@hospitaljobsin.com" {
		t.Errorf("Expected support email 'support@hospitaljobsin.com', got '%s'", templateData.SupportEmail)
	}

	templateData.SetField("custom_field", "custom_value")
	if templateData.GetField("custom_field") != "custom_value" {
		t.Error("Field set/get operations not working correctly")
	}

	dataMap := templateData.ToMap()
	if dataMap["app_name"] != templateData.AppName {
		t.Error("ToMap not including app_name correctly")
	}
	if dataMap["custom_field"] != "custom_value" {
		t.Error("ToMap not including custom fields correctly")
	}
}