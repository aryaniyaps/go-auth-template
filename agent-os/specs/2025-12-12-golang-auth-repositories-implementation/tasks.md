# Task Breakdown: Golang Auth Repositories Implementation

## Overview
Total Tasks: 29 tasks across 4 specialized task groups (consolidated approach with all repositories and errors in single repo.go file)

**NOTE:** Auth models are already implemented in `/server/internal/domain/auth/model.go`. This task list focuses only on pagination infrastructure and auth repository implementation.

## Task List

### Foundation Layer

#### Task Group 1: Pagination Infrastructure
**Dependencies:** None

- [x] 1.0 Complete pagination infrastructure
  - [x] 1.1 Write 2-8 focused tests for pagination system
    - Test cursor generation and parsing
    - Test PaginatedResult struct behavior
    - Test pagination edge cases (empty results, first/last boundaries)
    - Limit to 4-6 highly focused tests maximum
  - [x] 1.2 Create generic PaginatedResult[T] struct
    - Location: `/server/internal/infrastructure/db/pagination.go`
    - Fields: Data []T, HasNextPage bool, HasPreviousPage bool, StartCursor *string, EndCursor *string
    - Methods: NewPaginatedResult(), GetCursorForID()
  - [x] 1.3 Implement QueryBuilder extensions for pagination
    - Add ApplyCursorPagination() method to extend bun.DB queries
    - Support first, last, before, after parameters
    - Handle ID-based cursor encoding/decoding
  - [x] 1.4 Create pagination utilities package
    - Helper functions for cursor serialization
    - Constants for default page sizes
    - Error types for pagination failures
  - [x] 1.5 Ensure pagination tests pass
    - Run ONLY the 2-8 tests written in 1.1
    - Verify cursor-based pagination works correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 2-8 tests written in 1.1 pass
- PaginatedResult[T] works with any type
- Cursor-based pagination is performant and correct
- QueryBuilder extensions integrate seamlessly with Bun ORM

### Auth Repository Implementation

#### Task Group 2: Consolidated Auth Repository Implementation
**Dependencies:** Task Group 1

- [x] 2.0 Complete consolidated auth repositories
  - [x] 2.1 Write 2-8 focused tests for auth repositories
    - Test all 8 repository implementations (SessionRepo, PasswordResetTokenRepo, WebAuthnCredentialRepo, WebAuthnChallengeRepo, OAuthCredentialRepo, TwoFactorAuthenticationChallengeRepo, RecoveryCodeRepo, TemporaryTwoFactorChallengeRepo)
    - Test interface implementation satisfaction
    - Test method signature consistency
    - Test error wrapping and context preservation
    - Test static method availability
    - Limit to 6-8 focused tests maximum
  - [x] 2.2 Create consolidated auth repository file
    - Location: `/server/internal/domain/auth/repo.go`
    - Follow existing account repository pattern from `/server/internal/domain/account/repo.go`
    - Include all 8 repository implementations and custom errors in the same file
  - [x] 2.3 Define domain-specific errors in repo.go
    - Define: ErrSessionNotFound, ErrTokenExpired, ErrInvalidCredentials
    - Define: ErrWebAuthnCredentialNotFound, ErrChallengeNotFound, ErrRecoveryCodeInvalid
    - Use fmt.Errorf for context preservation following account repo pattern
  - [x] 2.4 Implement security utilities and static helpers
    - Add token generation and hashing functions (MD5 matching Python hashlib.md5)
    - Add crypto/rand token generation (matching secrets.token_hex)
    - Add static helper methods for each repository type:
      - generate_session_token(), hash_session_token()
      - generate_password_reset_token(), hash_password_reset_token()
      - generate_challenge(), hash_challenge() (for 2FA)
      - generate_recovery_code(), hash_recovery_code()
      - generate_two_factor_secret() (TOTP secret)
  - [x] 2.5 Implement all 8 repository structs in repo.go
    - SessionRepo: User session management with token hashing and pagination
    - PasswordResetTokenRepo: Password reset workflow with expiration
    - WebAuthnCredentialRepo: Security key management with sign count updates
    - WebAuthnChallengeRepo: Challenge generation and expiration
    - OAuthCredentialRepo: Provider account linking with uniqueness
    - TwoFactorAuthenticationChallengeRepo: TOTP secrets and challenge handling
    - RecoveryCodeRepo: Backup authentication code generation and validation
    - TemporaryTwoFactorChallengeRepo: Password reset 2FA workflow
  - [x] 2.6 Ensure consolidated tests pass
    - Run ONLY the 2-8 tests written in 2.1
    - Verify all repositories compile and satisfy interfaces
    - Verify error handling works across all scenarios
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 2-8 tests written in 2.1 pass
- All 8 repository implementations are properly defined in single file
- All custom errors are defined alongside repositories
- Method signatures match Python reference patterns
- Error wrapping preserves context and maintains consistency
- Static methods are correctly included in implementations
- Token generation and hashing match Python reference implementation

### Integration and Dependency Injection

#### Task Group 3: Integration and Configuration
**Dependencies:** Task Group 2

- [x] 3.0 Complete integration setup
  - [x] 3.1 Write 2-8 focused tests for dependency injection
    - Test repository initialization
    - Test database connectivity
    - Test lifecycle management
    - Limit to 2-4 focused tests maximum
  - [x] 3.2 Create repository provider/constructor
    - Location: `/server/internal/domain/auth/providers.go`
    - Use fx.Lifecycle patterns for dependency injection
    - Provide all 8 repositories as dependencies
    - Ensure proper database connection management
  - [x] 3.3 Set up database initialization
    - Create database indexes for performance
    - Set up foreign key constraints
    - Handle connection pooling configuration
  - [x] 3.4 Add graceful shutdown handling
    - Proper resource cleanup on shutdown
    - Database connection closure
    - In-flight operation completion
  - [x] 3.5 Ensure integration tests pass
    - Run ONLY the 2-8 tests written in 3.1
    - Verify dependency injection works correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- The 2-8 tests written in 3.1 pass
- All repositories are properly injectable
- Database connections are managed correctly
- Graceful shutdown works as expected

### Testing and Validation

#### Task Group 4: Test Review & Gap Analysis
**Dependencies:** Task Groups 1-3

- [x] 4.0 Review existing tests and fill critical gaps only
  - [x] 4.1 Review tests from Task Groups 1-3
    - Review the 2-8 tests from each task group (approximately 6-16 tests)
    - Identify patterns in test coverage across all repositories
    - Focus on test quality and strategic coverage
  - [x] 4.2 Analyze test coverage gaps for auth repositories only
    - Identify critical repository workflows that lack test coverage
    - Focus ONLY on gaps related to the 8 auth repositories
    - Prioritize integration scenarios between repositories
    - Skip edge cases and performance tests unless business-critical
  - [x] 4.3 Write up to 10 additional strategic tests maximum
    - Add maximum of 10 new tests to fill identified critical gaps
    - Focus on end-to-end repository workflows
    - Test pagination across different repository types
    - Test token lifecycle management
    - Do NOT write comprehensive coverage for all scenarios
    - Skip security penetration tests and performance benchmarks
  - [x] 4.4 Run auth-specific tests only
    - Run ONLY tests related to auth repositories (tests from all task groups and 4.3)
    - Expected total: approximately 16-26 tests maximum
    - Do NOT run the entire application test suite
    - Verify critical auth workflows pass

**Acceptance Criteria:**
- All auth-specific tests pass (approximately 16-26 tests total)
- Critical repository workflows for auth are covered
- No more than 10 additional tests added when filling in testing gaps
- Testing focused exclusively on auth repository requirements
- Cross-repository integration scenarios are tested

## Execution Order

Recommended implementation sequence:
1. **Pagination Infrastructure (Task Group 1):** Cursor-based pagination system
2. **Consolidated Auth Repository Implementation (Task Group 2):** All 8 repositories with errors and utilities in repo.go
3. **Integration and Dependency Injection (Task Group 3):** DI setup and configuration
4. **Testing and Validation (Task Group 4):** Test review and gap analysis

## Notes

- **Auth models already implemented** in `/server/internal/domain/auth/model.go` - this task list focuses only on pagination infrastructure and repository implementation
- **Consolidated approach:** All 8 auth repositories and custom errors are implemented in a single `/server/internal/domain/auth/repo.go` file
- Each task group follows a test-driven approach with 2-8 focused tests maximum
- Tests are limited to critical behaviors, not exhaustive coverage
- Task verification runs ONLY the newly written tests, not the entire suite
- Implementation follows existing account repository patterns from `/server/internal/domain/account/repo.go`
- Python reference implementation is used for behavior translation, not direct code copying
- Security considerations are prioritized throughout implementation
- Performance considerations (indexes, query optimization) are included in database operations
- **Total expected test count:** Approximately 16-26 tests maximum across all task groups