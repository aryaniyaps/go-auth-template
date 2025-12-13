# Specification: AccountService Implementation

## Goal
Port the Python AccountService from HospitalJobsIn to idiomatic Go, focusing on core account management functionality while leveraging existing Go patterns and infrastructure.

## User Stories
- As a developer, I want to port the Python AccountService to Go so that it integrates seamlessly with the existing Go auth template
- As a system architect, I want the Go implementation to follow idiomatic patterns and use dependency injection so that it maintains code quality and testability
- As a product owner, I want the Go version to maintain all core functionality including phone verification, analytics preferences, and S3 integration so that user experience remains consistent

## Specific Requirements

**Account Management Operations**
- Port GetAccountByPhoneNumber method with proper error handling and phone number validation
- Implement UpdateAccountFullName method with validation and database persistence
- Create UpdateAccountAvatarURL method with S3 integration using AWS SDK for Go
- Implement UpdateAccountPhoneNumber method with phone number format validation
- Port UpdateAccountTermsAndPolicy method for compliance tracking
- Port UpdateAccountAnalyticsPreference method for user privacy management
- Create UpdateAccountWhatsappJobAlerts method for notification preferences

**Phone Verification System**
- Implement CreatePhoneVerificationToken method with secure token generation
- Port VerifyPhoneNumber method with token validation and phone number verification
- Use appropriate Go SMS libraries for verification message sending
- Follow existing token generation and hashing patterns from account repo

**Message Integration**
- Reference HospitalJobsIn Python message abstraction for Go implementation
- Implement message sender interfaces for SMS notifications
- Use Go HTTP client libraries for external SMS service integration
- Follow async message patterns with proper error handling

**S3 Integration**
- Use AWS SDK for Go for avatar upload and deletion operations
- Implement proper S3 bucket configuration and security
- Handle file type validation and size limits
- Include error handling for S3 operation failures

**Dependency Injection Setup**
- Add AccountService to fx dependency injection system
- Integrate with existing AccountDomainModule
- Ensure proper constructor function signatures for fx.Provide
- Maintain separation of concerns with repository pattern

**Error Handling and Validation**
- Implement comprehensive error types matching existing patterns
- Add phone number validation using Go phonenumbers library
- Include input validation for all public methods
- Follow Go error wrapping best practices

**Testing Strategy**
- Create unit tests for all AccountService methods
- Mock external dependencies (S3, SMS services) for testing
- Include integration tests with database
- Test error scenarios and edge cases

## Existing Code to Leverage

**Account Repository (/home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/repo.go)**
- Complete AccountRepo interface with CRUD operations already implemented
- Token generation and hashing utilities (GenerateVerificationToken, HashVerificationToken)
- Phone number and email verification token repositories
- Error handling patterns and validation utilities
- Password hashing and verification functions

**Account Model (/home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/model.go)**
- Complete Account struct with all necessary fields including PhoneNumber, WhatsAppJobAlerts
- TermsAndPolicy and AnalyticsPreference embedded structs
- Verification token models with expiration checking
- Avatar URL generation helper method

**FX Dependency Injection Pattern (/home/aryaniyaps/go-projects/go-auth-template/server/internal/domain/account/providers.go)**
- AccountDomainModule pattern for dependency injection
- fx.Prove and fx.Options usage patterns
- Integration with main application setup

**Core Infrastructure (/home/aryaniyaps/go-projects/go-auth-template/server/cmd/server/main.go)**
- FX application setup and module integration
- Database and logger dependency patterns
- Existing domain module integration examples

## Out of Scope
- GraphQL resolvers or HTTP handlers (only service layer implementation)
- Database schema migrations (assume existing schema supports new fields)
- SMS provider configuration (focus on abstraction layer)
- S3 bucket setup and IAM policies (focus on SDK usage)
- Email verification functionality (already exists in account repo)
- Two-factor authentication implementation (already exists in account model)
- OAuth provider integrations (focus on core account operations only)