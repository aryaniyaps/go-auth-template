# Specification: GraphQL Authentication Context Injection

## Goal
Implement a Chi middleware that injects authentication context into GraphQL requests by validating JWE session tokens from HTTP-only cookies and providing account/session data to GraphQL resolvers.

## User Stories
- As a GraphQL developer, I want authentication context automatically available in resolvers so that I can implement authenticated operations without manual token parsing
- As a system administrator, I want session-based authentication with secure cookie storage so that user sessions are properly managed and secured
- As a user, I want my authentication state to persist across requests so that I don't need to re-authenticate for every operation

## Specific Requirements

**Session Management**
- Use JWE (JSON Web Encryption) tokens for secure session storage
- Store tokens in HTTP-only cookies to prevent XSS attacks
- Implement symmetric HS256 JWT algorithm for token signing
- Support fixed session expiration with database validation
- Use github.com/go-chi/jwtauth library for JWT handling

**Middleware Implementation**
- Create Chi middleware at `/server/internal/http/middleware/auth.go`
- Extract and validate session tokens from HTTP-only cookies
- Query database to validate session existence and expiration
- Set context values with keys "account" and "session" for GraphQL resolvers
- Set account to nil in context when no valid session found (simplified authentication check)
- Set session to nil in context when no valid session found
- Handle errors gracefully without interrupting request flow

**Database Integration**
- Leverage existing Session model from `/server/internal/domain/auth/model.go`
- Use existing Account model from `/server/internal/domain/account/model.go`
- Validate session against database with account relationship loading
- Respect session expiration timestamps stored in database

**GraphQL Handler Integration**
- Modify `AddGraphQLHandler` function in `/server/cmd/server/main.go`
- Apply authentication middleware before GraphQL handler
- Ensure context is available to all GraphQL resolvers
- Support existing GraphQL directives (@isAuthenticated, @requiresSudoMode)
- Simplify resolver authentication checks by testing for nil account

**Configuration and Dependencies**
- Add github.com/go-chi/jwtauth to go.mod dependencies
- Configure JWT secret through existing config system
- Support development and production environments
- Ensure middleware integrates with existing logging infrastructure

## Visual Design
No visual assets provided for this specification.

## Existing Code to Leverage

**Session Model** (`/server/internal/domain/auth/model.go`)
- Already defines Session struct with TokenHash, UserAgent, IPAddress, ExpiresAt, and AccountId fields
- Includes Account relationship for easy access to user data
- Has CoreModel integration for ID and timestamp fields

**Account Model** (`/server/internal/domain/account/model.go`)
- Complete Account model with authentication fields
- Supports 2FA, multiple auth providers, and user preferences
- Includes avatar URL generation and 2FA status methods

**Middleware Infrastructure** (`/server/internal/http/middleware/`)
- Existing logging middleware pattern to follow
- Chi router integration already established in `/server/internal/http/server.go`
- Middleware chain already configured with CORS, logging, and recovery

**GraphQL Schema** (`/server/graph/schema/auth.graphqls`)
- Session type already defined with proper GraphQL mappings
- Authentication directives (@isAuthenticated, @requiresSudoMode) ready for use
- Viewer query pattern for accessing authenticated user data

**HTTP Server Setup** (`/server/cmd/server/main.go`)
- AddGraphQLHandler function ready for middleware integration
- Fx dependency injection system for middleware registration
- Chi router with existing middleware chain

## Out of Scope
- Session cleanup and maintenance operations
- Cookie configuration for domain/path/security attributes
- Token refresh mechanisms and sliding expiration
- Logout endpoint implementation
- Multi-device session management
- Session revocation and forced logout
- CSRF protection implementation
- Rate limiting for authentication endpoints
- Session analytics and tracking
- OAuth integration and external provider sessions