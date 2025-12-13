# Spec Requirements: Account Service

## Initial Description
Implement a Go-based account management service similar to the Django account app from HospitalJobsIn. The service should handle user registration, phone verification, account management, analytics preferences, user blocking, and profile completion tracking.

## Requirements Discussion

### First Round Questions

**Q1:** I assume this service will handle both web-based user registration/login and API endpoints for mobile clients. Is that correct, or should we focus on just web-based authentication?

**Answer:** Both, but the priority is a robust API-based service first. Django templates won't be used, only JSON responses. The main focus is the `AccountService` which provides a comprehensive interface for user account operations.

**Q2:** For the message sending abstraction in Go, should we use interfaces to support multiple SMS providers (like in the Python implementation), or start with a specific provider like Twilio?

**Answer:** Use the AWS SDK for Go for S3 operations and use appropriate Go libraries for SMS functionality. Reference the message sending abstraction pattern from the HospitalJobsIn Python implementation at: https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/core/messages.py

**Q3:** I'm thinking we should use goroutines for background tasks like account deactivation cleanup or analytics processing. Should we implement this, or keep everything synchronous for the initial version?

**Answer:** Use goroutines only when required. Otherwise use synchronous Go.

**Q4:** Should we implement the complete integration with organizations, job alerts, and jobs domains in this initial version, or start with the core account functionality and stub out those external dependencies?

**Answer:** Only core account functionality. Stub out external dependencies to other domains like organizations, job alerts, and jobs.

### Existing Code to Reference
User provided reference to existing message sending abstraction pattern from HospitalJobsIn Python implementation.

**Similar Features Identified:**
- Message sending abstraction: `https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/core/messages.py`
- Go implementation should follow similar patterns for SMS sending

### Follow-up Questions
No follow-up questions were needed.

## Visual Assets

### Files Provided:
No visual assets provided.

### Visual Insights:
No visual insights available.

## Requirements Summary

### Functional Requirements
- **AccountService Implementation**: Focus on the main AccountService (lines 60-334 in the Django reference)
- **User Registration**: API-based user registration with phone verification
- **Account Management**: User profile management, email changes, password changes
- **Phone Verification**: Phone number verification using SMS
- **Analytics Preferences**: User consent management for analytics tracking
- **User Blocking**: Account blocking and unblocking functionality
- **Profile Completion Tracking**: Track and return profile completion status
- **JSON-only API**: No Django templates, only JSON responses
- **Cross-platform Support**: Support for both web and mobile clients

### Reusability Opportunities
- **Message Sending Pattern**: Reference the HospitalJobsIn Python message abstraction for Go implementation
- **AWS Integration**: Use AWS SDK for Go for S3 operations
- **SMS Libraries**: Use appropriate Go libraries for SMS functionality
- **Interface-based Design**: Follow the Python abstraction pattern for provider flexibility

### Scope Boundaries

**In Scope:**
- Core AccountService functionality
- User registration and authentication
- Phone verification system
- Account profile management
- Analytics preferences management
- User blocking/unblocking
- Profile completion tracking
- JSON API responses
- AWS S3 integration for file operations
- SMS integration for phone verification

**Out of Scope:**
- Django template rendering (JSON-only)
- Integration with organizations domain
- Integration with job alerts system
- Integration with jobs functionality
- Complex background processing (only required goroutines)
- Email templating system (beyond basic notification)

### Technical Considerations
- **Priority**: API-based service first, robust and scalable
- **Synchronous Approach**: Use synchronous Go unless goroutines are specifically required
- **AWS SDK**: Use AWS SDK for Go for cloud operations
- **SMS Libraries**: Use appropriate Go libraries for messaging
- **Interface Design**: Follow Python reference pattern for message abstraction
- **External Dependencies**: Stub out other domain services
- **File Path Reference**: Main focus on AccountService patterns from Django reference