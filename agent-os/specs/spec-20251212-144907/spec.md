# Specification: Authentication Repository Implementation

## Goal
Implement three focused repository interfaces (AccountRepo, EmailVerificationTokenRepo, PhoneNumberVerificationTokenRepo) using Go with Bun ORM, converting from Python MongoDB patterns while maintaining security best practices with argon2 password hashing and MD5 token hashing.

## User Stories
- As a developer, I want AccountRepo with CRUD and auth-specific methods so that I can manage user accounts with proper security
- As a developer, I want EmailVerificationTokenRepo with token lifecycle management so that I can handle email verification securely
- As a developer, I want PhoneNumberVerificationTokenRepo with token lifecycle management so that I can handle phone verification securely

## Specific Requirements

**Account Repository Interface and Implementation**
- Convert Python async methods to Go with context.Context support
- Create method: Create(ctx, email, fullName, authProviders, password, accountID, analyticsPreference, phoneNumber) (*Account, error)
- Get methods: Get(ctx, accountID) (*Account, error), GetByEmail(ctx, email) (*Account, error), GetByPhoneNumber(ctx, phone) (*Account, error)
- Update methods: Update(ctx, account, fullName, avatarURL, phoneNumber, termsAndPolicy, analyticsPreference, whatsappJobAlerts) (*Account, error)
- Profile management: UpdateProfile(ctx, account, profile) (*Account, error)
- Auth provider management: UpdateAuthProviders(ctx, account, authProviders) (*Account, error)
- 2FA management: SetTwoFactorSecret(ctx, account, totpSecret) (*Account, error), DeleteTwoFactorSecret(ctx, account) (*Account, error)
- Password management: UpdatePassword(ctx, account, password) (*Account, error), DeletePassword(ctx, account) (*Account, error)
- Avatar management: DeleteAvatar(ctx, account) (*Account, error)
- Deletion: Delete(ctx, account) error
- Static methods: HashPassword(password) string, VerifyPassword(password, hash) bool using argon2

**Python Source Code Reference:**
```python
class AccountRepo:
    async def create(
        self,
        email: str,
        full_name: str,
        auth_providers: list[AuthProvider],
        password: str | None = None,
        account_id: ObjectId | None = None,
        analytics_preference: Literal[
            "acceptance", "rejection", "undecided"
        ] = "undecided",
        phone_number: str | None = None,
    ) -> Account:
        """Create a new account."""
        account = Account(
            id=account_id,
            full_name=full_name,
            email=email,
            phone_number=phone_number,
            password_hash=self.hash_password(
                password=password,
            )
            if password is not None
            else None,
            updated_at=None,
            profile=None,
            auth_providers=auth_providers,
            terms_and_policy=TermsAndPolicy(
                type="acceptance",
                updated_at=datetime.now(UTC),
                version=TERMS_AND_POLICY_LATEST_VERSION,
            ),
            analytics_preference=AnalyticsPreference(
                type=analytics_preference,
                updated_at=datetime.now(UTC),
            ),
        )
        return await account.insert()

    @staticmethod
    def hash_password(password: str) -> str:
        return argon2.hash(password)

    @staticmethod
    def verify_password(password: str, password_hash: str) -> bool:
        return argon2.verify(password, password_hash)

    async def get(
        self,
        account_id: ObjectId,
        *,
        fetch_profile: bool = False,
    ) -> Account | None:
        """Get account by ID."""
        if fetch_profile:
            return await Account.get(
                account_id,
                fetch_links=True,
            )
        return await Account.get(account_id)

    async def get_by_email(self, email: str) -> Account | None:
        """Get account by email."""
        return await Account.find_one(
            Account.email == email,
        )

    async def get_by_phone_number(self, phone_number: str) -> Account | None:
        """Get account by phone number."""
        return await Account.find_one(
            Account.phone_number == phone_number,
        )

    async def update(
        self,
        account: Account,
        full_name: str = UNSET,
        avatar_url: str | None = UNSET,
        phone_number: str | None = UNSET,
        terms_and_policy: TermsAndPolicy = UNSET,
        analytics_preference: AnalyticsPreference = UNSET,
        whatsapp_job_alerts_enabled: bool = UNSET,
    ) -> Account:
        """Update the given account."""
        if full_name is not UNSET:
            account.full_name = full_name
        if avatar_url is not UNSET:
            account.avatar_url = avatar_url
        if phone_number is not UNSET:
            account.phone_number = phone_number
        if terms_and_policy is not UNSET:
            account.terms_and_policy = terms_and_policy
        if analytics_preference is not UNSET:
            account.analytics_preference = analytics_preference
        if whatsapp_job_alerts_enabled is not UNSET:
            account.whatsapp_job_alerts_enabled = whatsapp_job_alerts_enabled
        return await account.save()

    async def update_profile(self, account: Account, profile: Profile) -> Account:
        """Update the given account's profile."""
        account.profile = profile  # type: ignore[assignment]
        return await account.save()

    async def update_auth_providers(
        self,
        account: Account,
        auth_providers: list[AuthProvider],
    ) -> Account:
        """Update the given account."""
        account.auth_providers = auth_providers
        return await account.save()

    async def set_two_factor_secret(
        self, account: Account, totp_secret: str
    ) -> Account:
        """Set 2fa secret for the given account."""
        account.two_factor_secret = totp_secret
        return await account.save()

    async def delete_two_factor_secret(self, account: Account) -> Account:
        """Delete 2fa secret for the given account."""
        account.two_factor_secret = None
        return await account.save()

    async def update_password(self, account: Account, password: str) -> Account:
        """Update the given account's password."""
        if "password" not in account.auth_providers:
            # account initially had no password
            new_providers = account.auth_providers.copy()
            new_providers.append("password")
            account.auth_providers = new_providers
        account.password_hash = self.hash_password(password)
        return await account.save()

    async def delete_password(self, account: Account) -> Account:
        """Delete the given account's password."""
        account.password_hash = None
        if "password" in account.auth_providers:
            account.auth_providers.remove("password")
        return await account.save()

    async def delete_avatar(self, account: Account) -> Account:
        """Delete the given account's avatar."""
        account.avatar_url = None
        return await account.save()

    async def delete(self, account: Account) -> None:
        """Delete an account by ID."""
        await account.delete()
```

**Email Verification Token Repository Interface and Implementation**
- Create method: Create(ctx, email) (string, *EmailVerificationToken, error) returning generated token and stored entity
- Get methods: Get(ctx, verificationToken) (*EmailVerificationToken, error), GetByEmail(ctx, email) (*EmailVerificationToken, error)
- Deletion: Delete(ctx, emailVerification) error
- Static methods: GenerateVerificationToken() string, HashVerificationToken(token) string using MD5
- Token lookup by plaintext verification token with hash comparison

**Python Source Code Reference:**
```python
class EmailVerificationTokenRepo:
    async def create(self, email: str) -> str:
        """Create an email verification token."""
        verification_token = self.generate_verification_token()
        email_verification = EmailVerificationToken(
            email=email,
            token_hash=self.hash_verification_token(
                verification_token=verification_token,
            ),
            expires_at=datetime.now(UTC) + timedelta(hours=24),
        )
        await email_verification.insert()
        return verification_token

    async def get(self, verification_token: str) -> EmailVerificationToken | None:
        """Get an email verification token."""
        token_hash = self.hash_verification_token(
            verification_token=verification_token,
        )
        return await EmailVerificationToken.find_one(
            EmailVerificationToken.token_hash == token_hash
        )

    async def get_by_email(self, email: str) -> EmailVerificationToken | None:
        """Get an email verification token by email."""
        return await EmailVerificationToken.find_one(
            EmailVerificationToken.email == email,
        )

    async def delete(self, email_verification: EmailVerificationToken) -> None:
        """Delete an email verification token."""
        await email_verification.delete()

    @staticmethod
    def generate_verification_token(length: int = 32) -> str:
        """Generate a verification token."""
        return secrets.token_hex(length)

    @staticmethod
    def hash_verification_token(verification_token: str) -> str:
        """Hash a verification token."""
        return hashlib.md5(verification_token.encode()).hexdigest()
```

**Phone Number Verification Token Repository Interface and Implementation**
- Create method: Create(ctx, phoneNumber) (string, *PhoneNumberVerificationToken, error) returning generated token and stored entity
- Get methods: Get(ctx, verificationToken) (*PhoneNumberVerificationToken, error), GetByPhoneNumber(ctx, phoneNumber) (*PhoneNumberVerificationToken, error)
- Deletion: Delete(ctx, phoneNumberVerification) error
- Static methods: GenerateVerificationToken() string, HashVerificationToken(token) string using MD5
- Token lookup by plaintext verification token with hash comparison

**Python Source Code Reference:**
```python
class PhoneNumberVerificationTokenRepo:
    async def create(self, phone_number: str) -> str:
        """Create a phone number verification token."""
        verification_token = self.generate_verification_token()
        phone_number_verification = PhoneNumberVerificationToken(
            phone_number=phone_number,
            token_hash=self.hash_verification_token(
                verification_token=verification_token,
            ),
            expires_at=datetime.now(UTC) + timedelta(hours=24),
        )
        await phone_number_verification.insert()
        return verification_token

    async def get(self, verification_token: str) -> PhoneNumberVerificationToken | None:
        """Get a phone number verification token."""
        token_hash = self.hash_verification_token(
            verification_token=verification_token,
        )
        return await PhoneNumberVerificationToken.find_one(
            PhoneNumberVerificationToken.token_hash == token_hash
        )

    async def get_by_phone_number(
        self, phone_number: str
    ) -> PhoneNumberVerificationToken | None:
        """Get a phone number verification token by phone number."""
        return await PhoneNumberVerificationToken.find_one(
            PhoneNumberVerificationToken.phone_number == phone_number,
        )

    async def delete(
        self, phone_number_verification: PhoneNumberVerificationToken
    ) -> None:
        """Delete a phone number verification token."""
        await phone_number_verification.delete()

    @staticmethod
    def generate_verification_token(length: int = 32) -> str:
        """Generate a verification token."""
        return secrets.token_hex(length)

    @staticmethod
    def hash_verification_token(verification_token: str) -> str:
        """Hash a verification token."""
        return hashlib.md5(verification_token.encode()).hexdigest()
```

**Security Implementation**
- Use argon2 for password hashing (matching Python passlib.hash.argon2)
- Use MD5 for verification token hashing (matching Python hashlib.md5)
- Never store plaintext passwords or verification tokens
- Implement constant-time comparison for sensitive data verification
- Handle all hashing internally within repository methods
- Use existing TokenHash fields in verification token models

**Python Security Implementation Reference:**
```python
# Password hashing (from AccountRepo.hash_password method)
@staticmethod
def hash_password(password: str) -> str:
    return argon2.hash(password)

@staticmethod
def verify_password(password: str, password_hash: str) -> bool:
    return argon2.verify(password, password_hash)

# Token hashing (from verification token repositories)
@staticmethod
def hash_verification_token(verification_token: str) -> str:
    """Hash a verification token."""
    return hashlib.md5(verification_token.encode()).hexdigest()

@staticmethod
def generate_verification_token(length: int = 32) -> str:
    """Generate a verification token."""
    return secrets.token_hex(length)

# Token verification pattern
async def get(self, verification_token: str) -> EmailVerificationToken | None:
    """Get an email verification token."""
    token_hash = self.hash_verification_token(
        verification_token=verification_token,
    )
    return await EmailVerificationToken.find_one(
        EmailVerificationToken.token_hash == token_hash
    )
```

**Security Patterns Observed in Python Code:**
- Password hashing uses argon2 library for secure password storage
- Verification tokens use MD5 hashing (despite MD5 being cryptographically weak, it's used for token lookup, not security)
- Tokens are generated using secrets.token_hex() for cryptographically secure random values
- Hash comparison is done by database lookup rather than direct comparison
- All sensitive data (passwords, tokens) are hashed before storage
- Token expiration is enforced with ExpiresAt field and IsExpired() helper method

**Database Operations Pattern**
- Use *bun.DB from existing infrastructure/db/bun.go for database connections
- Follow existing Bun ORM patterns and tags from current models
- Leverage core.CoreModel for ID, CreatedAt, UpdatedAt fields
- Use proper Bun relationship loading for Account model relationships
- Handle unique constraint violations for email and phone_number fields
- Target file: /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/repo.go

**Error Handling Strategy**
- Use standard Go error types (sql.ErrNoRows for not found)
- Return descriptive errors for validation failures
- Handle duplicate email/phone constraints appropriately
- Use context-aware error handling with proper cancellation

**Method Signature Conversion**
- Convert Python async def to Go func with context.Context first parameter
- Convert Python type hints to Go types (list[AuthProvider] -> []string, ObjectId -> int64)
- Convert Python optional parameters (*string) to Go pointer types or special UNSET values
- Convert tuple returns to Go multiple return values
- Use Go naming conventions for methods and parameters

## Visual Design
No visual assets provided for this backend repository implementation.

## Existing Code to Leverage

**Account Model Structure**
- account.Account struct with FullName, Email, PasswordHash (*string), TwoFactorSecret (*string), AuthProviders ([]string)
- TermsAndPolicy and AnalyticsPreference embedded structs
- Existing helper methods: AvatarURL(), Has2FAEnabled(), TwoFactorProviders()

**Verification Token Models**
- account.EmailVerificationToken with Email, TokenHash, ExpiresAt fields
- account.PhoneNumberVerificationToken with PhoneNumber, TokenHash, ExpiresAt fields
- Both use core.CoreModel and have IsExpired() helper methods

**Database Infrastructure**
- *bun.DB connection setup from infrastructure/db/bun.go
- PostgreSQL dialect configuration with proper connection pooling
- Existing bun table aliases: accounts (acc), email_verification_tokens (evt), phone_verification_tokens (pvt)

**Core Model Foundation**
- core.CoreModel with ID (int64), CreatedAt, UpdatedAt fields
- Standardized bun tagging patterns across all models
- Autoincrement primary key and timestamp defaults

## Out of Scope
- Session repository implementation
- Password reset token repository
- WebAuthn credential repository
- OAuth credential repository
- Two-factor challenge repository
- Recovery code repository
- Authentication service layer business logic
- API endpoints, HTTP handlers, or middleware
- Database migration files or schema changes
- Token generation algorithms beyond basic random strings
- Email/SMS sending functionality
- 2FA TOTP code validation logic
- Transaction implementation (design to allow future addition)