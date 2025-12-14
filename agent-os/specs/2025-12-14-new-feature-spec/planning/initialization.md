# Feature Spec Initialization

## Feature Description
Implement GraphQL context injection middleware for authentication in go-auth-template. This involves setting up chi middleware to inject authenticated user context into GraphQL resolvers, enabling the existing @isAuthenticated and @requiresSudoMode directives to function properly.

## Initial Context
- **Date**: 2025-12-14
- **Project**: go-auth-template
- **Authentication Library**: github.com/go-chi/jwtauth
- **Session Management**: Database-backed sessions using existing bun models and repositories
- **GraphQL Schema**: Already has Session types, Account types, and authentication directives defined
- **Current State**: Schema has authentication directives but needs middleware implementation

## Technical Requirements
- Implement chi middleware for GraphQL context injection
- Use github.com/go-chi/jwtauth for JWT token handling
- Leverage existing Session bun model and repository
- Follow the pattern established in the hospitaljobsin project's dependencies.py and sessions.py
- Enable @isAuthenticated and @requiresSudoMode directives to work properly

## Next Steps
This feature focuses specifically on the middleware implementation for context injection, not schema changes or mutation implementation.