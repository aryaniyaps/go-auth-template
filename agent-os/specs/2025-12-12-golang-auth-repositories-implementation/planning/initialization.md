# Golang Auth Repositories Implementation

## Initial Description

This spec is for implementing auth repositories in Golang using Bun ORM and PostgreSQL, transitioning from a Python reference implementation that uses MongoDB and Beanie ODM.

### Reference Implementation
The reference implementation is located at:
https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/auth/repositories.py

### Context
This is for a Go authentication template project. The implementation needs to translate the Python authentication repository patterns to Go, maintaining similar functionality but adapting to:

- **Language**: Python → Go
- **Database**: MongoDB → PostgreSQL
- **ORM**: Beanie ODM → Bun ORM
- **Auth Components**: Session management, password reset, WebAuthn, OAuth, 2FA challenges, recovery codes

### Repository Types to Implement
Based on the reference code, the following repositories need to be implemented:

1. **SessionRepo** - User session management
2. **PasswordResetTokenRepo** - Password reset token handling
3. **WebAuthnCredentialRepo** - WebAuthn credential management
4. **WebAuthnChallengeRepo** - WebAuthn challenge handling
5. **OauthCredentialRepo** - OAuth credential storage
6. **TwoFactorAuthenticationChallengeRepo** - 2FA challenge management
7. **RecoveryCodeRepo** - Recovery code generation and validation
8. **TemporaryTwoFactorChallengeRepo** - Temporary 2FA challenges

### Key Requirements
- Maintain similar API patterns and functionality
- Adapt Go-specific patterns and error handling
- Use Bun ORM for PostgreSQL operations
- Implement proper security practices (token hashing, expiration, etc.)
- Support pagination where applicable
- Include proper transaction handling