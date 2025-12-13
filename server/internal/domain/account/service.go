package account

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ttacon/libphonenumber"
	"go.uber.org/zap"
)

const (
	MaxAvatarFileSize   = 5 << 20 // 5MB
	AllowedAvatarTypes  = "image/jpeg,image/png,image/gif,image/webp"
	AvatarBucketName    = "account-avatars"
	AvatarURLExpiry     = 24 * time.Hour
	SMSTokenLength      = 6
	SMSTokenExpiry      = 15 * time.Minute
)

var (
	phoneNumberRegex = regexp.MustCompile(`^\+\d{10,15}$`)
)

// AccountService provides business logic for account operations
type AccountService struct {
	accountRepo              AccountRepo
	phoneTokenRepo           PhoneNumberVerificationTokenRepo
	emailTokenRepo           EmailVerificationTokenRepo
	messageSender            MessageSender
	s3Client                 *s3.Client
	logger                   *zap.Logger
}

// NewAccountService creates a new AccountService instance
func NewAccountService(
	accountRepo AccountRepo,
	phoneTokenRepo PhoneNumberVerificationTokenRepo,
	emailTokenRepo EmailVerificationTokenRepo,
	messageSender MessageSender,
	s3Client *s3.Client, // Optional dependency
	logger *zap.Logger,
) *AccountService {
	return &AccountService{
		accountRepo:    accountRepo,
		phoneTokenRepo: phoneTokenRepo,
		emailTokenRepo: emailTokenRepo,
		messageSender:  messageSender,
		s3Client:       s3Client, // Can be nil
		logger:         logger,
	}
}

// NewS3Client creates a new S3 client with default configuration
func NewS3Client(ctx context.Context, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return s3.NewFromConfig(cfg), nil
}

// NewS3ClientFromEnv creates an S3 client using environment variables
func NewS3ClientFromEnv(ctx context.Context) (*s3.Client, error) {
	region := "us-east-1" // Default region
	if envRegion := ctx.Value("aws_region"); envRegion != nil {
		if r, ok := envRegion.(string); ok {
			region = r
		}
	}
	return NewS3Client(ctx, region)
}

// GetAccountByPhoneNumber retrieves an account by phone number
//
// This method validates the phone number format using international phone number standards
// and then retrieves the associated account from the repository.
//
// Parameters:
//   - ctx: Context for the request
//   - phoneNumber: Phone number in international format (e.g., "+1234567890")
//
// Returns:
//   - *Account: The account associated with the phone number
//   - error: ErrAccountNotFound if no account exists, or other errors for validation/database issues
//
// Example:
//   account, err := service.GetAccountByPhoneNumber(ctx, "+1234567890")
//   if err != nil {
//       log.Printf("Failed to get account: %v", err)
//       return
//   }
//   fmt.Printf("Found account: %s\n", account.FullName)
func (s *AccountService) GetAccountByPhoneNumber(ctx context.Context, phoneNumber string) (*Account, error) {
	if err := s.validatePhoneNumber(phoneNumber); err != nil {
		return nil, err
	}

	account, err := s.accountRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, ErrAccountNotFound
		}
		s.logger.Error("Failed to get account by phone number", zap.Error(err))
		return nil, fmt.Errorf("failed to get account by phone number: %w", err)
	}

	return account, nil
}

// UpdateAccountFullName updates an account's full name
//
// This method validates the input full name, retrieves the existing account,
// and updates it with the new full name.
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - fullName: New full name for the account (non-empty string)
//
// Returns:
//   - *Account: The updated account
//   - error: Error if account not found, validation fails, or database update fails
//
// Example:
//   account, err := service.UpdateAccountFullName(ctx, accountID, "John Doe")
//   if err != nil {
//       log.Printf("Failed to update full name: %v", err)
//       return
//   }
//   fmt.Printf("Updated account: %s\n", account.FullName)
func (s *AccountService) UpdateAccountFullName(ctx context.Context, accountID int64, fullName string) (*Account, error) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return nil, fmt.Errorf("full name cannot be empty")
	}

	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account, &fullName, nil, nil, nil, nil, nil)
	if err != nil {
		s.logger.Error("Failed to update account full name", zap.Error(err))
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return updatedAccount, nil
}

// UpdateAccountPhoneNumber updates an account's phone number
//
// This method validates the phone number format, retrieves the existing account,
// and updates it with the new phone number. Phone numbers must be in international format.
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - phoneNumber: New phone number in international format (e.g., "+1234567890")
//
// Returns:
//   - *Account: The updated account
//   - error: Error if account not found, phone validation fails, or database update fails
//
// Example:
//   account, err := service.UpdateAccountPhoneNumber(ctx, accountID, "+1234567890")
//   if err != nil {
//       log.Printf("Failed to update phone number: %v", err)
//       return
//   }
//   fmt.Printf("Updated phone number: %s\n", account.PhoneNumber)
func (s *AccountService) UpdateAccountPhoneNumber(ctx context.Context, accountID int64, phoneNumber string) (*Account, error) {
	if err := s.validatePhoneNumber(phoneNumber); err != nil {
		return nil, err
	}

	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account, nil, nil, &phoneNumber, nil, nil, nil)
	if err != nil {
		s.logger.Error("Failed to update account phone number", zap.Error(err))
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return updatedAccount, nil
}

// UpdateAccountTermsAndPolicy updates an account's terms and policy acceptance
//
// This method updates the account's terms and policy acceptance status with a version
// number and timestamp. This is useful for tracking compliance with different versions
// of terms of service and privacy policies.
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - version: Version string of the terms and policy being accepted (e.g., "2.1.0")
//
// Returns:
//   - *Account: The updated account
//   - error: Error if account not found, validation fails, or database update fails
//
// Example:
//   account, err := service.UpdateAccountTermsAndPolicy(ctx, accountID, "2.1.0")
//   if err != nil {
//       log.Printf("Failed to update terms: %v", err)
//       return
//   }
//   fmt.Printf("Accepted terms version: %s\n", account.TermsAndPolicy.Version)
func (s *AccountService) UpdateAccountTermsAndPolicy(ctx context.Context, accountID int64, version string) (*Account, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("invalid input: terms version cannot be empty")
	}

	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	termsAndPolicy := &TermsAndPolicy{
		Type:      "accepted",
		Version:   version,
		UpdatedAt: time.Now(),
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account, nil, nil, nil, termsAndPolicy, nil, nil)
	if err != nil {
		s.logger.Error("Failed to update account terms and policy", zap.Error(err))
		return nil, fmt.Errorf("failed to update account terms and policy: %w", err)
	}

	return updatedAccount, nil
}

// UpdateAccountAnalyticsPreference updates an account's analytics preference
//
// This method updates the account's analytics preference, which controls whether
// the user's data can be used for analytics purposes. Valid values are "enabled",
// "disabled", or "undecided".
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - preference: Analytics preference - "enabled", "disabled", or "undecided"
//
// Returns:
//   - *Account: The updated account
//   - error: Error if account not found, validation fails, or database update fails
//
// Example:
//   account, err := service.UpdateAccountAnalyticsPreference(ctx, accountID, "enabled")
//   if err != nil {
//       log.Printf("Failed to update analytics preference: %v", err)
//       return
//   }
//   fmt.Printf("Analytics preference: %s\n", account.AnalyticsPreference.Type)
func (s *AccountService) UpdateAccountAnalyticsPreference(ctx context.Context, accountID int64, preference string) (*Account, error) {
	preference = strings.TrimSpace(preference)
	validPreferences := []string{"enabled", "disabled", "undecided"}

	isValid := false
	for _, valid := range validPreferences {
		if preference == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, fmt.Errorf("invalid preference value: preference must be 'enabled', 'disabled', or 'undecided'")
	}

	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	analyticsPref := &AnalyticsPreference{
		Type:      preference,
		UpdatedAt: time.Now(),
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account, nil, nil, nil, nil, analyticsPref, nil)
	if err != nil {
		s.logger.Error("Failed to update account analytics preference", zap.Error(err))
		return nil, fmt.Errorf("failed to update account analytics preference: %w", err)
	}

	return updatedAccount, nil
}

// UpdateAccountWhatsappJobAlerts updates an account's WhatsApp job alerts preference
//
// This method enables or disables WhatsApp job alerts for the user's account.
// When enabled, the user will receive job notifications via WhatsApp.
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - enabled: Whether to enable (true) or disable (false) WhatsApp job alerts
//
// Returns:
//   - *Account: The updated account
//   - error: Error if account not found or database update fails
//
// Example:
//   account, err := service.UpdateAccountWhatsappJobAlerts(ctx, accountID, true)
//   if err != nil {
//       log.Printf("Failed to update WhatsApp alerts: %v", err)
//       return
//   }
//   fmt.Printf("WhatsApp alerts enabled: %t\n", account.WhatsappJobAlerts)
func (s *AccountService) UpdateAccountWhatsappJobAlerts(ctx context.Context, accountID int64, enabled bool) (*Account, error) {
	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	updatedAccount, err := s.accountRepo.Update(ctx, account, nil, nil, nil, nil, nil, &enabled)
	if err != nil {
		s.logger.Error("Failed to update account WhatsApp job alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to update account WhatsApp job alerts: %w", err)
	}

	return updatedAccount, nil
}

// UpdateAccountAvatarURL updates an account's avatar URL by uploading file to S3
//
// This method handles the complete avatar upload workflow including file validation,
// S3 upload, and account update. The file is validated for type and size before upload.
// Requires S3 client to be configured.
//
// Parameters:
//   - ctx: Context for the request
//   - accountID: ID of the account to update
//   - file: File content reader (e.g., from multipart form)
//   - filename: Original filename of the file being uploaded
//
// Returns:
//   - *Account: The updated account with new avatar URL
//   - error: Error if validation fails, S3 upload fails, or database update fails
//
// Example:
//   file, header, err := r.FormFile("avatar")
//   if err != nil {
//       log.Printf("Failed to get file: %v", err)
//       return
//   }
//   defer file.Close()
//
//   account, err := service.UpdateAccountAvatarURL(ctx, accountID, file, header.Filename)
//   if err != nil {
//       log.Printf("Failed to upload avatar: %v", err)
//       return
//   }
//   fmt.Printf("Avatar uploaded: %s\n", account.InternalAvatarURL)
func (s *AccountService) UpdateAccountAvatarURL(ctx context.Context, accountID int64, file io.Reader, filename string) (*Account, error) {
	// Validate inputs first
	if file == nil {
		return nil, fmt.Errorf("invalid input: file cannot be nil")
	}

	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, fmt.Errorf("invalid input: filename cannot be empty")
	}

	// Check if S3 client is configured
	if s.s3Client == nil {
		return nil, fmt.Errorf("S3 client not configured")
	}

	// Get account
	account, err := s.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Validate file
	fileBytes, contentType, err := s.validateAvatarFile(file, filename)
	if err != nil {
		return nil, err
	}

	// Generate unique filename
	uniqueFilename := s.generateUniqueFilename(filename)

	// Upload to S3
	avatarURL, err := s.uploadToS3(ctx, uniqueFilename, fileBytes, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update account with new avatar URL
	updatedAccount, err := s.accountRepo.Update(ctx, account, nil, &avatarURL, nil, nil, nil, nil)
	if err != nil {
		s.logger.Error("Failed to update account avatar URL", zap.Error(err))
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return updatedAccount, nil
}

// validateAvatarFile validates the uploaded avatar file
func (s *AccountService) validateAvatarFile(file io.Reader, filename string) ([]byte, string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check file size
	if len(fileBytes) == 0 {
		return nil, "", fmt.Errorf("file is empty")
	}
	if len(fileBytes) > MaxAvatarFileSize {
		return nil, "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxAvatarFileSize)
	}

	// Detect content type
	contentType := http.DetectContentType(fileBytes)

	// Parse allowed types
	allowedTypesList := strings.Split(AllowedAvatarTypes, ",")
	isAllowed := false
	for _, allowedType := range allowedTypesList {
		if strings.TrimSpace(allowedType) == contentType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return nil, "", fmt.Errorf("file type %s is not allowed", contentType)
	}

	return fileBytes, contentType, nil
}

// generateUniqueFilename generates a unique filename for S3 upload
func (s *AccountService) generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d%s", timestamp, ext)
	return uniqueFilename
}

// uploadToS3 uploads file content to S3 and returns the URL
func (s *AccountService) uploadToS3(ctx context.Context, filename string, fileBytes []byte, contentType string) (string, error) {
	// Create PutObject input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(AvatarBucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Make the file publicly accessible
	}

	// Upload file
	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct URL
	avatarURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", AvatarBucketName, filename)

	s.logger.Info("Successfully uploaded avatar to S3",
		zap.String("bucket", AvatarBucketName),
		zap.String("filename", filename),
		zap.String("url", avatarURL))

	return avatarURL, nil
}

// CreatePhoneVerificationToken creates and sends a phone verification token
//
// This method generates a 6-digit verification token, stores it in the database,
// and sends it to the user's phone number via SMS. The token expires after 15 minutes.
//
// Parameters:
//   - ctx: Context for the request
//   - phoneNumber: Phone number to send the verification token to (international format)
//
// Returns:
//   - error: Error if phone validation fails, token creation fails, or SMS sending fails
//
// Example:
//   err := service.CreatePhoneVerificationToken(ctx, "+1234567890")
//   if err != nil {
//       log.Printf("Failed to create verification token: %v", err)
//       return
//   }
//   fmt.Println("Verification token sent successfully")
func (s *AccountService) CreatePhoneVerificationToken(ctx context.Context, phoneNumber string) error {
	if err := s.validatePhoneNumber(phoneNumber); err != nil {
		return fmt.Errorf("invalid phone number format: %w", err)
	}

	// Create verification token
	token, phoneToken, err := s.phoneTokenRepo.Create(ctx, phoneNumber)
	if err != nil {
		s.logger.Error("Failed to create phone verification token", zap.Error(err))
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send SMS with token
	message := fmt.Sprintf("Your verification code is: %s. This code will expire in 15 minutes.", token)
	err = s.messageSender.SendSMS(ctx, phoneNumber, message)
	if err != nil {
		s.logger.Error("Failed to send SMS verification", zap.Error(err))
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	s.logger.Info("Phone verification token created and sent",
		zap.String("phone_number", phoneNumber),
		zap.Int64("token_id", phoneToken.ID))

	return nil
}

// VerifyPhoneNumber verifies a phone number using the provided token
//
// This method validates the provided token against the stored hash and marks the
// phone number as verified. If an account exists with this phone number, it will
// be updated with the verified phone number. The verification token is deleted
// after successful verification.
//
// Parameters:
//   - ctx: Context for the request
//   - phoneNumber: Phone number being verified (international format)
//   - token: 6-digit verification token received via SMS
//
// Returns:
//   - error: Error if validation fails, token is invalid/expired, or database operations fail
//
// Example:
//   err := service.VerifyPhoneNumber(ctx, "+1234567890", "123456")
//   if err != nil {
//       log.Printf("Failed to verify phone number: %v", err)
//       return
//   }
//   fmt.Println("Phone number verified successfully")
func (s *AccountService) VerifyPhoneNumber(ctx context.Context, phoneNumber, token string) error {
	phoneNumber = strings.TrimSpace(phoneNumber)
	token = strings.TrimSpace(token)

	if phoneNumber == "" {
		return fmt.Errorf("invalid input: phone number cannot be empty")
	}
	if token == "" {
		return fmt.Errorf("invalid input: token cannot be empty")
	}

	if err := s.validatePhoneNumber(phoneNumber); err != nil {
		return fmt.Errorf("invalid phone number format: %w", err)
	}

	// Get verification token
	phoneToken, err := s.phoneTokenRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		if err == ErrTokenNotFound {
			return fmt.Errorf("invalid verification token: no verification token found for this phone number")
		}
		s.logger.Error("Failed to get phone verification token", zap.Error(err))
		return fmt.Errorf("failed to get verification token: %w", err)
	}

	// Check if token has expired
	if time.Now().After(phoneToken.ExpiresAt) {
		s.logger.Warn("Verification token has expired",
			zap.String("phone_number", phoneNumber),
			zap.Time("expired_at", phoneToken.ExpiresAt))
		return fmt.Errorf("token has expired")
	}

	// Verify token hash
	tokenHash := s.phoneTokenRepo.HashVerificationToken(token)
	if tokenHash != phoneToken.TokenHash {
		s.logger.Warn("Invalid verification token provided",
			zap.String("phone_number", phoneNumber),
			zap.Int64("token_id", phoneToken.ID))
		return fmt.Errorf("invalid verification token")
	}

	// Check if account exists with this phone number
	account, err := s.accountRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil && err != ErrAccountNotFound {
		s.logger.Error("Failed to check account by phone number", zap.Error(err))
		return fmt.Errorf("failed to verify phone number: %w", err)
	}

	// If account exists, update it with verified phone number
	if account != nil {
		_, err = s.accountRepo.Update(ctx, account, nil, nil, &phoneNumber, nil, nil, nil)
		if err != nil {
			s.logger.Error("Failed to update account with verified phone number", zap.Error(err))
			return fmt.Errorf("failed to update account: %w", err)
		}
	}

	// Delete the used token
	err = s.phoneTokenRepo.Delete(ctx, phoneToken)
	if err != nil {
		// Log error but don't fail the verification
		s.logger.Warn("Failed to delete used verification token",
			zap.String("phone_number", phoneNumber),
			zap.Int64("token_id", phoneToken.ID),
			zap.Error(err))
	}

	s.logger.Info("Phone number verified successfully",
		zap.String("phone_number", phoneNumber),
		zap.Bool("account_updated", account != nil))

	return nil
}

// validatePhoneNumber validates phone number format
func (s *AccountService) validatePhoneNumber(phoneNumber string) error {
	phoneNumber = strings.TrimSpace(phoneNumber)

	if phoneNumber == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Use libphonenumber for comprehensive validation
	parsed, err := libphonenumber.Parse(phoneNumber, "")
	if err != nil {
		return fmt.Errorf("invalid phone number format: %w", err)
	}

	if !libphonenumber.IsValidNumber(parsed) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}