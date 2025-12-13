# AccountService Porting Requirements

## Python Source Code Reference

### Source Repository
- **URL**: https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/accounts/services.py
- **Lines to Port**: 60-334 (AccountService class only)
- **License**: Refer to repository license terms

### Message Abstraction Reference
- **URL**: https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/core/messages.py
- **Purpose**: Understand message sending patterns for Go implementation

## Detailed Method Breakdown

### Core AccountService Methods to Port

#### 1. GetAccountByPhoneNumber
**Python Implementation**:
```python
def get_account_by_phone_number(self, phone_number: str) -> Optional[Account]:
    """Get account by phone number"""
    return self.db.query(Account).filter(Account.phone_number == phone_number).first()
```

**Go Implementation Requirements**:
- Use existing `accountRepo.GetByPhoneNumber()` method
- Add phone number validation using phonenumbers library
- Handle `ErrAccountNotFound` appropriately
- Return proper error types

#### 2. UpdateAccountFullName
**Python Implementation**:
```python
def update_account_full_name(self, account_id: int, full_name: str) -> Account:
    """Update account full name"""
    account = self.get_account_by_id(account_id)
    account.full_name = full_name
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use existing `accountRepo.Update()` method with fullName pointer
- Add input validation for full_name
- Handle update errors appropriately
- Return updated account with proper error handling

#### 3. UpdateAccountAvatarURL (S3 Integration)
**Python Implementation**:
```python
def update_account_avatar_url(self, account_id: int, avatar_url: str) -> Account:
    """Update account avatar URL"""
    account = self.get_account_by_id(account_id)
    account.avatar_url = avatar_url
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use AWS SDK for Go for S3 operations
- Implement file upload to S3 with proper error handling
- Update account with S3 URL using `accountRepo.Update()`
- Handle S3 bucket configuration and permissions
- Include file type validation and size limits

#### 4. UpdateAccountPhoneNumber
**Python Implementation**:
```python
def update_account_phone_number(self, account_id: int, phone_number: str) -> Account:
    """Update account phone number"""
    account = self.get_account_by_id(account_id)
    account.phone_number = phone_number
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use existing `accountRepo.Update()` method with phoneNumber pointer
- Add phone number format validation using phonenumbers library
- Handle unique constraint violations for phone numbers
- Return updated account with proper error handling

#### 5. UpdateAccountTermsAndPolicy
**Python Implementation**:
```python
def update_account_terms_and_policy(self, account_id: int, terms_version: str) -> Account:
    """Update account terms and policy acceptance"""
    account = self.get_account_by_id(account_id)
    account.terms_and_policy = {
        "type": "accepted",
        "updated_at": datetime.utcnow(),
        "version": terms_version
    }
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use existing `accountRepo.Update()` method with TermsAndPolicy pointer
- Create TermsAndPolicy struct with proper timestamp
- Handle version tracking for policy compliance
- Return updated account with proper error handling

#### 6. UpdateAccountAnalyticsPreference
**Python Implementation**:
```python
def update_account_analytics_preference(self, account_id: int, preference: str) -> Account:
    """Update account analytics preference"""
    account = self.get_account_by_id(account_id)
    account.analytics_preference = {
        "type": preference,
        "updated_at": datetime.utcnow()
    }
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use existing `accountRepo.Update()` method with AnalyticsPreference pointer
- Validate preference values ("enabled", "disabled", "undecided")
- Create AnalyticsPreference struct with proper timestamp
- Return updated account with proper error handling

#### 7. UpdateAccountWhatsappJobAlerts
**Python Implementation**:
```python
def update_account_whatsapp_job_alerts(self, account_id: int, enabled: bool) -> Account:
    """Update account WhatsApp job alerts preference"""
    account = self.get_account_by_id(account_id)
    account.whatsapp_job_alerts = enabled
    self.db.commit()
    self.db.refresh(account)
    return account
```

**Go Implementation Requirements**:
- Use existing `accountRepo.Update()` method with whatsappJobAlerts pointer
- Handle boolean preference updates
- Return updated account with proper error handling

### Phone Verification Methods

#### 8. CreatePhoneVerificationToken
**Python Implementation**:
```python
def create_phone_verification_token(self, phone_number: str) -> str:
    """Create phone verification token and send SMS"""
    # Generate 6-digit code
    token = str(random.randint(100000, 999999))

    # Hash and store token
    hashed_token = hashlib.sha256(token.encode()).hexdigest()
    verification_token = PhoneNumberVerificationToken(
        phone_number=phone_number,
        token_hash=hashed_token,
        expires_at=datetime.utcnow() + timedelta(hours=24)
    )
    self.db.add(verification_token)
    self.db.commit()

    # Send SMS
    self.send_verification_sms(phone_number, token)
    return token
```

**Go Implementation Requirements**:
- Use existing `phoneNumberVerificationTokenRepo.Create()` method
- Generate 6-digit verification code
- Hash token using existing `HashVerificationToken()` utility
- Send SMS using message abstraction implementation
- Handle SMS sending failures appropriately
- Return generated token (for testing purposes)

#### 9. VerifyPhoneNumber
**Python Implementation**:
```python
def verify_phone_number(self, phone_number: str, token: str) -> bool:
    """Verify phone number with token"""
    # Find token
    verification_token = self.db.query(PhoneNumberVerificationToken).filter(
        PhoneNumberVerificationToken.phone_number == phone_number,
        PhoneNumberVerificationToken.expires_at > datetime.utcnow()
    ).first()

    if not verification_token:
        return False

    # Verify token
    hashed_token = hashlib.sha256(token.encode()).hexdigest()
    if verification_token.token_hash != hashed_token:
        return False

    # Update account phone number
    account = self.get_account_by_id(verification_token.account_id)
    account.phone_number = phone_number
    account.phone_number_verified = True

    # Delete token
    self.db.delete(verification_token)
    self.db.commit()

    return True
```

**Go Implementation Requirements**:
- Use existing `phoneNumberVerificationTokenRepo.GetByPhoneNumber()` method
- Verify token hash using existing `HashVerificationToken()` utility
- Check token expiration
- Update account phone number using `accountRepo.Update()`
- Delete verification token using `phoneNumberVerificationTokenRepo.Delete()`
- Return boolean success indicator

### Message Integration Requirements

#### SMS Message Sender
**Python Reference**:
```python
from app.core.messages import BaseMessageSender, SMSMessageSender

class AccountService:
    def __init__(self, db: Session, message_sender: BaseMessageSender):
        self.db = db
        self.message_sender = message_sender

    def send_verification_sms(self, phone_number: str, token: str):
        message = f"Your verification code is: {token}"
        self.message_sender.send_sms(phone_number, message)
```

**Go Implementation Requirements**:
- Define MessageSender interface following BaseMessageSender pattern
- Implement SMSMessageSender struct with Go HTTP client
- Use external SMS provider API (Twilio, AWS SNS, etc.)
- Include proper error handling and retry logic
- Support for international phone number formats

## Go Implementation Patterns

### Service Structure
```go
type AccountService struct {
    accountRepo                        AccountRepo
    phoneVerificationTokenRepo         PhoneNumberVerificationTokenRepo
    emailVerificationTokenRepo         EmailVerificationTokenRepo
    messageSender                      MessageSender
    s3Client                          *s3.Client
    logger                            *zap.Logger
}

func NewAccountService(
    accountRepo AccountRepo,
    phoneVerificationTokenRepo PhoneNumberVerificationTokenRepo,
    emailVerificationTokenRepo EmailVerificationTokenRepo,
    messageSender MessageSender,
    s3Client *s3.Client,
    logger *zap.Logger,
) *AccountService {
    return &AccountService{
        accountRepo:                 accountRepo,
        phoneVerificationTokenRepo:  phoneVerificationTokenRepo,
        emailVerificationTokenRepo:  emailVerificationTokenRepo,
        messageSender:              messageSender,
        s3Client:                   s3Client,
        logger:                     logger,
    }
}
```

### Error Handling Pattern
```go
var (
    ErrInvalidPhoneNumber   = errors.New("invalid phone number format")
    ErrTokenExpired        = errors.New("verification token has expired")
    ErrInvalidToken        = errors.New("invalid verification token")
    ErrSMSSendFailed       = errors.New("failed to send SMS verification")
    ErrS3UploadFailed      = errors.New("failed to upload file to S3")
)

func (s *AccountService) UpdateAccountFullName(ctx context.Context, accountID int64, fullName string) (*Account, error) {
    if strings.TrimSpace(fullName) == "" {
        return nil, fmt.Errorf("%w: full name cannot be empty", ErrInvalidInput)
    }

    account, err := s.accountRepo.Get(ctx, accountID)
    if err != nil {
        return nil, fmt.Errorf("failed to get account: %w", err)
    }

    return s.accountRepo.Update(ctx, account, &fullName, nil, nil, nil, nil, nil)
}
```

### S3 Integration Pattern
```go
func (s *AccountService) UpdateAccountAvatarURL(ctx context.Context, accountID int64, file io.Reader, filename string) (*Account, error) {
    // Validate file type and size
    if err := s.validateAvatarFile(file, filename); err != nil {
        return nil, fmt.Errorf("invalid avatar file: %w", err)
    }

    // Upload to S3
    key := fmt.Sprintf("avatars/%d/%s", accountID, filename)
    _, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.config.S3Bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(getContentType(filename)),
        ACL:         types.ObjectCannedACLPrivate,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to upload avatar to S3: %w", err)
    }

    // Generate S3 URL
    avatarURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.config.S3Bucket, key)

    // Update account
    account, err := s.accountRepo.Get(ctx, accountID)
    if err != nil {
        return nil, fmt.Errorf("failed to get account: %w", err)
    }

    return s.accountRepo.Update(ctx, account, nil, &avatarURL, nil, nil, nil, nil)
}
```

### FX Integration
```go
// Add to /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/providers.go
var AccountDomainModule = fx.Options(
    fx.Provide(
        NewAccountRepo,
        NewEmailVerificationTokenRepo,
        NewPhoneNumberVerificationTokenRepo,
        NewAccountService,
        NewMessageSender,
        NewS3Client,
    ),
)
```

## Dependencies to Add

### Go Modules Required
```go
require (
    github.com/aws/aws-sdk-go-v2 v1.x.x
    github.com/aws/aws-sdk-go-v2/service/s3 v1.x.x
    github.com/aws/aws-sdk-go-v2/config v1.x.x
    github.com/nyaruka/phonenumbers v1.x.x  // Phone number validation
    github.com/twilio/twilio-go v1.x.x     // SMS service (or alternative)
)
```

### Configuration Requirements
- S3 bucket configuration (bucket name, region, access keys)
- SMS provider configuration (API keys, from number)
- Phone number validation patterns
- File upload limits and allowed types

## Testing Requirements

### Unit Tests
- Test all AccountService methods with mocked dependencies
- Test error scenarios and edge cases
- Test phone number validation logic
- Test S3 upload operations with mocked S3 client
- Test SMS sending with mocked message sender

### Integration Tests
- Test database operations with real database (test environment)
- Test S3 operations with test bucket
- Test SMS integration with test provider

### Test Structure Example
```go
func TestAccountService_UpdateAccountFullName(t *testing.T) {
    tests := []struct {
        name         string
        accountID    int64
        fullName     string
        setupMocks   func(*MockAccountRepo)
        expectError  bool
        errorMessage string
    }{
        {
            name:      "successful update",
            accountID: 1,
            fullName:  "John Doe",
            setupMocks: func(m *MockAccountRepo) {
                m.On("Get", mock.Anything, int64(1)).Return(&Account{}, nil)
                m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&Account{}, nil)
            },
            expectError: false,
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Implementation Checklist

### Phase 1: Core Service Setup
- [ ] Create AccountService struct with all required dependencies
- [ ] Implement NewAccountService constructor function
- [ ] Add AccountService to FX dependency injection
- [ ] Set up basic logging and error handling

### Phase 2: Account Management Methods
- [ ] Implement GetAccountByPhoneNumber method
- [ ] Implement UpdateAccountFullName method
- [ ] Implement UpdateAccountPhoneNumber method
- [ ] Implement UpdateAccountTermsAndPolicy method
- [ ] Implement UpdateAccountAnalyticsPreference method
- [ ] Implement UpdateAccountWhatsappJobAlerts method

### Phase 3: S3 Integration
- [ ] Set up AWS S3 client configuration
- [ ] Implement file validation helper functions
- [ ] Implement UpdateAccountAvatarURL with S3 upload
- [ ] Add S3 error handling and cleanup

### Phase 4: Phone Verification System
- [ ] Implement CreatePhoneVerificationToken method
- [ ] Implement VerifyPhoneNumber method
- [ ] Set up SMS message sender interface and implementation
- [ ] Add phone number validation using phonenumbers library

### Phase 5: Testing and Documentation
- [ ] Write comprehensive unit tests for all methods
- [ ] Write integration tests for database and S3 operations
- [ ] Add method documentation and examples
- [ ] Performance testing and optimization

### Phase 6: Integration and Deployment
- [ ] Integration testing with existing auth template
- [ ] Update main.go to include AccountService module
- [ ] Environment configuration setup
- [ ] Monitoring and logging setup