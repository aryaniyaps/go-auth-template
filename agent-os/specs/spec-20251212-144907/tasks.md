# Task Breakdown: Authentication Repository Implementation

## Overview
Total Tasks: 32

This implementation focuses on converting Python MongoDB repository patterns to Go PostgreSQL with Bun ORM, implementing secure authentication repositories with proper password hashing (argon2) and token hashing (MD5).

## Task List

### Core Infrastructure & Security Setup

#### Task Group 1: Foundation and Security Utilities
**Dependencies:** None

- [x] 1.0 Complete foundation and security utilities
  - [x] 1.1 Write 2-4 focused tests for security utilities
    - Test password hashing and verification functions
    - Test token generation and hashing functions
    - Test constant-time comparison if implemented
  - [x] 1.2 Implement argon2 password hashing utilities
    - HashPassword(password string) (string, error)
    - VerifyPassword(password, hash string) (bool, error)
    - Use golang.org/x/crypto/argon2 for compatibility with Python passlib
    - Match Python argon2 parameters exactly
  - [x] 1.3 Implement MD5 token hashing utilities
    - HashVerificationToken(token string) string using crypto/md5
    - GenerateVerificationToken(length int) (string, error) using crypto/rand
    - Ensure cryptographically secure token generation
  - [x] 1.4 Create helper function for updating array fields in models
    - Handle AuthProviders []string updates with proper slice management
    - Reuse pattern for other array field updates
  - [x] 1.5 Ensure security utilities tests pass
    - Run ONLY the 2-4 tests written in 1.1
    - Verify argon2 compatibility with Python implementation
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-4 tests written in 1.1 pass
- [x] Password hashing is compatible with Python argon2
- [x] Token generation is cryptographically secure
- [x] MD5 hashing matches Python hashlib.md5 output

### Account Repository Implementation

#### Task Group 2: Account Repository Interface and Basic CRUD
**Dependencies:** Task Group 1

- [x] 2.0 Complete Account repository interface and CRUD operations
  - [x] 2.1 Write 2-4 focused tests for basic Account CRUD
    - Test Create method with all field combinations
    - Test Get methods (by ID, email, phone)
    - Test Update method with field validation
  - [x] 2.2 Define AccountRepo interface
    - All required method signatures with context.Context first parameter
    - Proper Go return types: (*Account, error), error, etc.
    - Follow Go naming conventions and error handling
  - [x] 2.3 Implement AccountRepo struct with *bun.DB dependency
    - Constructor: NewAccountRepo(db *bun.DB) *AccountRepo
    - Embed *bun.DB field for database operations
  - [x] 2.4 Create method implementation
    - Create(ctx, email, fullName, authProviders, password, accountID, analyticsPreference, phoneNumber) (*Account, error)
    - Handle optional parameters with pointer types
    - Hash password using utilities from Task Group 1
    - Handle unique constraint violations for email/phone
    - Set default TermsAndPolicy and AnalyticsPreference
  - [x] 2.5 Get methods implementation
    - Get(ctx, accountID) (*Account, error) handling sql.ErrNoRows
    - GetByEmail(ctx, email) (*Account, error)
    - GetByPhoneNumber(ctx, phone) (*Account, error)
    - Use Bun ORM patterns with proper WHERE clauses
  - [x] 2.6 Ensure Account CRUD tests pass
    - Run ONLY the 2-4 tests written in 2.1
    - Verify database operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-4 tests written in 2.1 pass
- [x] Account creation works with all field combinations
- [x] Get methods properly handle not found cases
- [x] Database constraints are enforced correctly

#### Task Group 3: Account Repository Update Operations
**Dependencies:** Task Group 2

- [x] 3.0 Complete Account repository update operations
  - [x] 3.1 Write 2-4 focused tests for Account update methods
    - Test Update method with different field combinations
    - Test profile and auth provider updates
    - Test avatar and deletion operations
  - [x] 3.2 Implement Update method
    - Update(ctx, account, fullName, avatarURL, phoneNumber, termsAndPolicy, analyticsPreference, whatsappJobAlerts) (*Account, error)
    - Handle optional parameters with pointer types or UNSET pattern
    - Update only provided fields, preserve others unchanged
    - Use Bun's NewUpdate() with WHERE clause for safety
  - [x] 3.3 Implement UpdateProfile method
    - UpdateProfile(ctx, account, profile) (*Account, error)
    - Handle profile data structure and embedding
    - Use proper Bun relationship loading if needed
  - [x] 3.4 Implement UpdateAuthProviders method
    - UpdateAuthProviders(ctx, account, authProviders) (*Account, error)
    - Handle []string field updates with array tagging
    - Validate auth provider values if required
  - [x] 3.5 Implement DeleteAvatar method
    - DeleteAvatar(ctx, account) (*Account, error)
    - Set avatar_url field to NULL
    - Use Bun's Set() method for nullable field updates
  - [x] 3.6 Ensure Account update tests pass
    - Run ONLY the 2-4 tests written in 3.1
    - Verify all update operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-4 tests written in 3.1 pass
- [x] Update methods preserve unchanged fields correctly
- [x] Array field updates work with Bun ORM tagging
- [x] Nullable field updates handle NULL values properly

#### Task Group 4: Account Repository Security Operations
**Dependencies:** Task Group 3

- [x] 4.0 Complete Account repository security operations
  - [x] 4.1 Write 2-4 focused tests for security operations
    - Test password hash/verify static methods
    - Test 2FA secret management
    - Test password update and deletion
  - [x] 4.2 Implement static password methods
    - HashPassword(password) string and VerifyPassword(password, hash) bool
    - Reuse from Task Group 1 but as static methods on AccountRepo
    - Ensure thread-safe operation (argon2 is safe for concurrent use)
  - [x] 4.3 Implement 2FA management methods
    - SetTwoFactorSecret(ctx, account, totpSecret) (*Account, error)
    - DeleteTwoFactorSecret(ctx, account) (*Account, error)
    - Handle nullable TwoFactorSecret field properly
  - [x] 4.4 Implement password management methods
    - UpdatePassword(ctx, account, password) (*Account, error)
    - Add "password" to AuthProviders if not already present
    - DeletePassword(ctx, account) (*Account, error)
    - Remove "password" from AuthProviders when deleting password
  - [x] 4.5 Implement Delete method
    - Delete(ctx, account) error
    - Use Bun's NewDelete() with WHERE clause for safety
    - Handle cascading deletes if required by schema
  - [x] 4.6 Ensure Account security tests pass
    - Run ONLY the 2-4 tests written in 4.1
    - Verify security operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-4 tests written in 4.1 pass
- [x] Password hashing matches Python argon2 implementation
- [x] 2FA secret management handles nullable fields correctly
- [x] AuthProviders list updates work for password management

### Email Verification Token Repository

#### Task Group 5: Email Verification Token Repository
**Dependencies:** Task Group 1

- [x] 5.0 Complete Email Verification Token repository
  - [x] 5.1 Write 2-3 focused tests for EmailVerificationTokenRepo
    - Test Create method with token generation
    - Test Get methods (by token, by email)
    - Test Delete method
  - [x] 5.2 Define EmailVerificationTokenRepo interface
    - All required method signatures with context.Context
    - Proper return types matching specification
    - Follow Go naming conventions
  - [x] 5.3 Implement EmailVerificationTokenRepo struct
    - NewEmailVerificationTokenRepo(db *bun.DB) *EmailVerificationTokenRepo
    - Embed *bun.DB field for database operations
  - [x] 5.4 Implement Create method
    - Create(ctx, email) (string, *EmailVerificationToken, error)
    - Generate token and hash using Task Group 1 utilities
    - Set expires_at to 24 hours from creation
    - Return plaintext token and stored entity
  - [x] 5.5 Implement Get methods
    - Get(ctx, verificationToken) (*EmailVerificationToken, error)
    - Hash incoming token and compare with stored TokenHash
    - GetByEmail(ctx, email) (*EmailVerificationToken, error)
    - Handle expired token checking using IsExpired() method
  - [x] 5.6 Implement Delete method
    - Delete(ctx, emailVerification) error
    - Use Bun's NewDelete() with primary key
  - [x] 5.7 Ensure EmailVerificationTokenRepo tests pass
    - Run ONLY the 2-3 tests written in 5.1
    - Verify token lifecycle operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-3 tests written in 5.1 pass
- [x] Token generation is cryptographically secure
- [x] Token lookup by plaintext token works correctly
- [x] Expiration handling works with IsExpired() method

### Phone Number Verification Token Repository

#### Task Group 6: Phone Number Verification Token Repository
**Dependencies:** Task Group 1

- [x] 6.0 Complete Phone Number Verification Token repository
  - [x] 6.1 Write 2-3 focused tests for PhoneNumberVerificationTokenRepo
    - Test Create method with token generation
    - Test Get methods (by token, by phone number)
    - Test Delete method
  - [x] 6.2 Define PhoneNumberVerificationTokenRepo interface
    - All required method signatures with context.Context
    - Proper return types matching specification
    - Follow Go naming conventions
  - [x] 6.3 Implement PhoneNumberVerificationTokenRepo struct
    - NewPhoneNumberVerificationTokenRepo(db *bun.DB) *PhoneNumberVerificationTokenRepo
    - Embed *bun.DB field for database operations
  - [x] 6.4 Implement Create method
    - Create(ctx, phoneNumber) (string, *PhoneNumberVerificationToken, error)
    - Generate token and hash using Task Group 1 utilities
    - Set expires_at to 24 hours from creation
    - Return plaintext token and stored entity
  - [x] 6.5 Implement Get methods
    - Get(ctx, verificationToken) (*PhoneNumberVerificationToken, error)
    - Hash incoming token and compare with stored TokenHash
    - GetByPhoneNumber(ctx, phoneNumber) (*PhoneNumberVerificationToken, error)
    - Handle expired token checking using IsExpired() method
  - [x] 6.6 Implement Delete method
    - Delete(ctx, phoneNumberVerification) error
    - Use Bun's NewDelete() with primary key
  - [x] 6.7 Ensure PhoneNumberVerificationTokenRepo tests pass
    - Run ONLY the 2-3 tests written in 6.1
    - Verify token lifecycle operations work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-3 tests written in 6.1 pass
- [x] Token generation is cryptographically secure
- [x] Token lookup by plaintext token works correctly
- [x] Phone number handling works with unique constraints

### Integration and Error Handling

#### Task Group 7: Error Handling and Integration
**Dependencies:** Task Groups 1-6

- [x] 7.0 Complete error handling and integration
  - [x] 7.1 Write 2-4 focused integration tests
    - Test cross-repository operations if needed
    - Test error handling scenarios
    - Test concurrent operations if applicable
  - [x] 7.2 Implement proper error handling patterns
    - Use sql.ErrNoRows for not found scenarios
    - Create descriptive error types for validation failures
    - Handle unique constraint violations properly
    - Use context-aware error handling with cancellation
  - [x] 7.3 Add repository constructor functions
    - NewAccountRepo(db *bun.DB) *AccountRepo
    - NewEmailVerificationTokenRepo(db *bun.DB) *EmailVerificationTokenRepo
    - NewPhoneNumberVerificationTokenRepo(db *bun.DB) *PhoneNumberVerificationTokenRepo
  - [x] 7.4 Verify all repositories work together
    - Test database transaction compatibility
    - Test with existing *bun.DB infrastructure
    - Verify proper connection handling
  - [x] 7.5 Ensure integration tests pass
    - Run ONLY the 2-4 tests written in 7.1
    - Verify error handling works correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- [x] The 2-4 tests written in 7.1 pass
- [x] Error handling is consistent across all repositories
- [x] Database operations integrate properly with existing infrastructure
- [x] Context cancellation works correctly

### Testing and Verification

#### Task Group 8: Test Review & Gap Analysis
**Dependencies:** Task Groups 1-7

- [x] 8.0 Review existing tests and fill critical gaps only
  - [x] 8.1 Review tests from Task Groups 1-7
    - Review the 2-4 tests from foundation utilities (Task 1.1)
    - Review the 2-4 tests from Account CRUD (Task 2.1)
    - Review the 2-4 tests from Account updates (Task 3.1)
    - Review the 2-4 tests from Account security (Task 4.1)
    - Review the 2-3 tests from EmailVerificationTokenRepo (Task 5.1)
    - Review the 2-3 tests from PhoneNumberVerificationTokenRepo (Task 6.1)
    - Review the 2-4 tests from integration (Task 7.1)
    - Total existing tests: approximately 14-26 tests
  - [x] 8.2 Analyze test coverage gaps for authentication repositories only
    - Identify critical repository workflows lacking test coverage
    - Focus ONLY on gaps related to repository functionality
    - Do NOT assess entire application test coverage
    - Prioritize end-to-end repository workflows over unit test gaps
  - [x] 8.3 Write up to 8 additional strategic tests maximum
    - Add maximum of 8 new tests to fill identified critical gaps
    - Focus on integration between repositories and database
    - Focus on error handling and edge cases in repository operations
    - Do NOT write comprehensive coverage for all scenarios
    - Skip performance tests and load testing unless business-critical
  - [x] 8.4 Run repository-specific tests only
    - Run ONLY tests related to authentication repositories
    - Expected total: approximately 22-34 tests maximum
    - Do NOT run the entire application test suite
    - Verify critical repository workflows pass

**Acceptance Criteria:**
- [x] All repository-specific tests pass (31 tests total)
- [x] Critical repository workflows are covered
- [x] No more than 8 additional tests added when filling in testing gaps
- [x] Testing focused exclusively on authentication repository requirements

## Execution Order

Recommended implementation sequence:
1. [x] Foundation and Security Utilities (Task Group 1)
2. [x] Account Repository Interface and Basic CRUD (Task Group 2)
3. [x] Account Repository Update Operations (Task Group 3)
4. [x] Account Repository Security Operations (Task Group 4)
5. [x] Email Verification Token Repository (Task Group 5)
6. [x] Phone Number Verification Token Repository (Task Group 6)
7. [x] Error Handling and Integration (Task Group 7)
8. [x] Test Review & Gap Analysis (Task Group 8)

## Key Implementation Notes

### Security Requirements
- [x] Use argon2 for password hashing with parameters matching Python passlib
- [x] Use MD5 for verification token hashing (matching Python hashlib.md5)
- [x] Never store plaintext passwords or verification tokens
- [x] Implement proper constant-time comparison where applicable
- [x] All sensitive data must be hashed before storage

### Database Patterns
- [x] Use existing *bun.DB infrastructure from server/internal/infrastructure/db/bun.go
- [x] Follow existing Bun ORM patterns with proper table aliases (accounts: acc, email_verification_tokens: evt, phone_verification_tokens: pvt)
- [x] Leverage core.CoreModel for ID, CreatedAt, UpdatedAt fields
- [x] Handle unique constraint violations for email and phone_number fields
- [x] Use context.Context for all database operations with proper cancellation

### Go Patterns
- [x] Follow Go naming conventions for methods and parameters
- [x] Use proper error handling with descriptive error types
- [x] Convert Python async patterns to Go with context.Context support
- [x] Handle optional parameters with pointer types or special UNSET values
- [x] Use Go multiple return values instead of Python tuples

## Current Test Coverage Summary
- **Task Group 1:** 6 tests (security utilities)
- **Task Group 2:** 4 tests (account CRUD)
- **Task Group 3:** 3 tests (account updates)
- **Task Group 4:** 8 tests (account security operations)
- **Task Group 5:** 4 tests (email verification tokens)
- **Task Group 6:** 4 tests (phone verification tokens)
- **Task Group 7:** 2 tests (integration)
- **Total Current Tests:** 31 tests