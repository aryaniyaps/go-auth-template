# Spec Requirements: Authentication Repository Implementation

## Initial Description
Implement authentication repository layer for the Go auth template using Bun ORM. The repository layer should provide CRUD operations for Account, EmailVerificationToken, PhoneNumberVerificationToken, Session, PasswordResetToken, and all other auth-related models with proper password hashing, transaction support, and error handling.

## Requirements Discussion

### First Round Questions

**Q1:** What authentication-related Bun models are available in the existing codebase?
**Answer:**
- Account model exists in /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/model.go with fields: FullName, Email, PasswordHash, TwoFactorSecret, InternalAvatarURL, AuthProviders, TermsAndPolicy, AnalyticsPref
- EmailVerificationToken model exists with Email, TokenHash, ExpiresAt (time.Time)
- PhoneNumberVerificationToken model exists with PhoneNumber, TokenHash, ExpiresAt (time.Time)
- All models use core.CoreModel (includes ID, CreatedAt, UpdatedAt)
- Additional auth models in /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/auth/model.go: Session, PasswordResetToken, WebAuthnCredential, WebAuthnChallenge, OAuthCredential, TwoFactorAuthenticationChallenge, RecoveryCode, TemporaryTwoFactorChallenge

**Q2:** How should we ensure consistency with existing database patterns in the codebase?
**Answer:**
- Auth models use TokenHash (never plaintext tokens)
- Some use int64 timestamps (Session.ExpiresAt), others use time.Time (verification tokens)
- Account relationships use Account *account.Account with bun relationship tags
- Verification tokens use unique constraints on email/phone_number
- Follow existing bun tagging patterns and use core.CoreModel for base fields

**Q3:** What specific repository methods should we implement for each model based on the Python reference implementation?
**Answer:**
Based on Python implementation, need core CRUD plus:
- AccountRepo: Create, GetByID, GetByEmail, GetByPhoneNumber, Update, UpdatePassword, Set2FA, Delete2FA, Delete
- EmailVerificationTokenRepo: Create, GetByToken, GetByEmail, Delete
- PhoneNumberVerificationTokenRepo: Create, GetByToken, GetByPhoneNumber, Delete
- Plus repositories for all additional auth models (Session, PasswordResetToken, WebAuthnCredential, etc.)

**Q4:** How should password hashing be handled in the Go implementation?
**Answer:**
- Account model has PasswordHash *string (nullable for OAuth)
- Should use bcrypt (standard in Go)
- Repository should handle hashing internally (matches Python pattern)

**Q5:** What timestamp approach should we use for consistency across the auth models?
**Answer:**
- Verification tokens already use time.Time - keep this consistent
- No need to standardize on int64 for these specific repos
- Session model already uses int64 for ExpiresAt

**Q6:** What transaction support is needed?
**Answer:**
- Yes, support context.Context
- No explicit transaction requirements mentioned, but design to allow future addition

**Q7:** What error handling approach should we use?
**Answer:**
- Use standard Go errors initially
- Can add custom error types later if needed

### Existing Code to Reference

**Similar Features Identified:**
- Account model: /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/model.go
- Auth models: /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/auth/model.go
- Empty repository file: /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/auth/repo.go
- Database setup: /home/aryaniyaps/go-projects/go-auth-template/server/internal/infrastructure/db/bun.go
- Core model: /home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/core/model.go

**Components to potentially reuse:**
- bun relationship tags and patterns from existing models
- core.CoreModel for base fields
- database connection setup patterns from bun.go
- existing model validation and helper methods

**Backend logic to reference:**
- Account.AvatarURL() method for avatar generation
- Account.Has2FAEnabled() and Account.TwoFactorProviders() methods
- Email/Phone verification token IsExpired() methods

### Follow-up Questions
No follow-up questions were needed - user provided comprehensive answers.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
None - no design mockups or wireframes were provided.

## Requirements Summary

### Functional Requirements
- Implement repository layer for all auth-related models using Bun ORM
- Provide CRUD operations with auth-specific methods (GetByEmail, GetByToken, etc.)
- Handle password hashing internally using bcrypt
- Support context.Context for all operations
- Follow existing bun tagging and relationship patterns
- Design for future transaction support

### Reusability Opportunities
- Use existing bun relationship patterns from Account model
- Follow core.CoreModel structure used across domain models
- Reference existing model validation and helper methods
- Build on database connection setup from infrastructure/db/bun.go

### Scope Boundaries
**In Scope:**
- Repository interfaces and implementations for Account model
- Repository interfaces and implementations for EmailVerificationToken
- Repository interfaces and implementations for PhoneNumberVerificationToken
- Repository interfaces and implementations for all additional auth models (Session, PasswordResetToken, WebAuthnCredential, WebAuthnChallenge, OAuthCredential, TwoFactorAuthenticationChallenge, RecoveryCode, TemporaryTwoFactorChallenge)
- Password hashing using bcrypt
- Basic CRUD operations plus auth-specific query methods
- Context support for all operations

**Out of Scope:**
- Custom error types (use standard Go errors initially)
- Explicit transaction implementation (design to allow future addition)
- Authentication service layer (only repositories)
- API endpoints or handlers
- Migration files or database schema changes
- Token generation or validation logic (only storage/retrieval)

### Technical Considerations
- Integration points: Use existing *bun.DB from infrastructure/db/bun.go
- Existing system constraints: Follow current domain structure and naming conventions
- Technology preferences: Bun ORM, PostgreSQL, bcrypt for password hashing
- Similar code patterns to follow: Use bun relationship tags, core.CoreModel embedding, and existing model validation patterns