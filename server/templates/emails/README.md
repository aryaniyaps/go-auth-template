# Email Templates

This directory contains email templates converted from Jinja2 to Pongo2 syntax for the HospitalJobs authentication system.

## Template Structure

```
templates/emails/
├── base/                    # Base templates (currently not used due to Pongo2 limitations)
│   ├── body.mjml
│   └── body.txt
├── email-verification/      # Email verification templates
│   ├── body.mjml           # HTML version with MJML
│   ├── body.txt            # Plain text version
│   └── subject.txt         # Email subject line
└── password-reset/         # Password reset templates
    ├── body.mjml           # HTML version with MJML
    ├── body.txt            # Plain text version
    └── subject.txt         # Email subject line
```

## Available Templates

### Email Verification Templates

**Purpose:** Send email verification codes to users for email address verification.

**Files:**
- `email-verification/body.mjml` - HTML email with verification code display
- `email-verification/body.txt` - Plain text version
- `email-verification/subject.txt` - Email subject

**Required Variables:**
- `app_name` - Application name
- `app_url` - Application URL
- `email` - User's email address
- `verification_token` - Email verification code
- `token_expires_in` - Token expiration time (e.g., "15 minutes")
- `user_agent` - User's browser user agent
- `support_email` - Support email address

**Usage:**
```go
data := EmailVerificationData(cfg, "user@example.com", "123456", "Mozilla/5.0")
err := emailClient.SendEmailTemplate(ctx, "emails/email-verification", data, []string{"user@example.com"})
```

### Password Reset Templates

**Purpose:** Send password reset links for both initial password setup and password resets.

**Files:**
- `password-reset/body.mjml` - HTML email with reset button
- `password-reset/body.txt` - Plain text version
- `password-reset/subject.txt` - Email subject with conditional logic

**Required Variables:**
- `app_name` - Application name
- `app_url` - Application URL
- `reset_link` - Password reset URL
- `link_expires_in` - Link expiration time (e.g., "1 hour")
- `is_initial` - Boolean: true for initial setup, false for password reset
- `user_agent` - User's browser user agent
- `support_email` - Support email address

**Conditional Logic:**
The templates use `{% if is_initial %}` to differentiate between:
- Initial password setup (`is_initial: true`)
- Password reset (`is_initial: false`)

**Usage:**
```go
// Initial password setup
data := PasswordResetData(cfg, "https://example.com/reset?token=abc", "Mozilla/5.0", true)
err := emailClient.SendEmailTemplate(ctx, "emails/password-reset", data, []string{"user@example.com"})

// Password reset
data := PasswordResetData(cfg, "https://example.com/reset?token=xyz", "Mozilla/5.0", false)
err := emailClient.SendEmailTemplate(ctx, "emails/password-reset", data, []string{"user@example.com"})
```

## Template Syntax (Pongo2)

The templates use Pongo2 syntax, which is compatible with Jinja2 for most common operations:

### Variable Output
```pongo2
{{ variable_name }}
```

### Conditionals
```pongo2
{% if condition %}
    Content when condition is true
{% else %}
    Content when condition is false
{% endif %}
```

### Variable Assignment
```pongo2
{% set action_text = "reset your password" %}
```

### Filters
```pongo2
{{ variable|capitalize }}
```

### MJML Integration
MJML templates are processed first by Pongo2 for variable substitution, then by MJML for HTML rendering.

## Configuration

Templates are configured via environment variables:

```bash
# Template directory path
EMAIL_TEMPLATE_PATH=./templates/emails

# Email provider (dummy, smtp, ses, sendgrid)
EMAIL_PROVIDER=dummy

# Common email settings
FROM_EMAIL=noreply@hospitaljobsin.com
FROM_NAME=HospitalJobs
```

## Helper Functions

The email package provides convenience functions for template data:

### EmailVerificationData
```go
data := EmailVerificationData(cfg, "user@example.com", "123456", "Mozilla/5.0")
```

### PasswordResetData
```go
data := PasswordResetData(cfg, "reset-url", "Mozilla/5.0", false)
```

### Manual Template Creation
```go
templateData := NewEmailTemplateData(cfg)
templateData.SetField("custom_field", "custom_value")
data := templateData.ToMap()
```

## Testing

Templates are tested with the following scenarios:

1. **Variable Substitution** - All template variables render correctly
2. **Conditional Logic** - Password reset templates handle initial vs reset scenarios
3. **MJML Rendering** - MJML templates generate valid HTML
4. **Data Helpers** - Template data functions generate correct data structures
5. **Email Client Integration** - Templates work with the email client

Run tests:
```bash
go test ./internal/infrastructure/email -v
```

## Customization

### Adding New Templates

1. Create a new directory under `templates/emails/`
2. Add `body.mjml`, `body.txt`, and `subject.txt` files
3. Use Pongo2 syntax for variables and logic
4. Create helper functions for template data if needed
5. Add tests for the new templates

### Modifying Existing Templates

1. Edit the template files directly
2. Test changes with the provided test suite
3. Update helper functions if new variables are needed
4. Update documentation if required

### Customizing Application Data

The default application data can be customized in `helpers.go`:

```go
data := &EmailTemplateData{
    AppName:      "Your App Name",
    AppURL:       "https://yourapp.com",
    SupportEmail: "support@yourapp.com",
    Data:         make(map[string]interface{}),
}
```

Or make it configurable by adding fields to the main Config struct.

## Base Templates Note

Due to Pongo2 limitations with template inheritance using `{% extends %}`, the current templates are self-contained rather than using base templates. If template inheritance is needed in the future, consider:

1. Using a different templating engine with better inheritance support
2. Implementing custom template composition logic
3. Creating macro-based templates instead of inheritance

## Best Practices

1. **Always provide both HTML and text versions** of emails for accessibility
2. **Use meaningful variable names** that clearly indicate their purpose
3. **Test templates with real data** to ensure they render correctly
4. **Keep templates responsive** by using MJML best practices
5. **Validate email addresses** before sending emails
6. **Handle edge cases** in template variables (empty strings, null values, etc.)
7. **Use consistent styling** across all email templates
8. **Include unsubscribe links** when required by regulations
9. **Test emails across different email clients** to ensure compatibility
10. **Use appropriate security measures** for verification tokens and reset links

## Troubleshooting

### Template Not Found Errors
Ensure the `EMAIL_TEMPLATE_PATH` environment variable points to the correct directory and that template files exist with proper permissions.

### Variable Not Rendering
Check that:
- Variable names match between template and data
- Data is passed in the correct format (map[string]interface{})
- No syntax errors in template

### MJML Rendering Issues
- Validate MJML syntax using online MJML validators
- Check for unclosed MJML tags
- Ensure template variables don't break MJML structure

### Email Not Sending
- Verify email provider configuration
- Check provider credentials
- Test with dummy provider first
- Check email client initialization