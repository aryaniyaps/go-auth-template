# Specification: Golang Auth Repositories Implementation

## Goal
Implement 8 comprehensive auth repositories using interface-first pattern with Bun ORM and PostgreSQL, translating Python/MongoDB patterns to idiomatic Go with cursor-based pagination and domain-specific error handling.

## User Stories
- As a developer, I want to use consistent auth repositories across all authentication components so that I can maintain clean architecture and reusable patterns
- As a system administrator, I want secure token handling and proper session management so that user authentication remains reliable and performant
- As a user, I want seamless authentication experiences across multiple auth providers so that I can access the system using my preferred method

## Specific Requirements

**Repository Interface Architecture**
- Create interface-first pattern following existing AccountRepo structure in `/server/internal/domain/account/repo.go`
- Define all 8 repository interfaces: SessionRepo, PasswordResetTokenRepo, WebAuthnCredentialRepo, WebAuthnChallengeRepo, OauthCredentialRepo, TwoFactorAuthenticationChallengeRepo, RecoveryCodeRepo, TemporaryTwoFactorChallengeRepo
- Include static helper methods within interfaces for token generation, hashing, and validation
- Implement repository struct patterns with Bun DB injection following existing accountRepo pattern

**Cursor-Based Pagination System**
- Create generic PaginatedResult[T] struct in `/server/internal/infrastructure/db/pagination.go`
- Implement cursor-based pagination using ID-based cursors for performance
- Abstract pagination logic to internal/infrastructure/db package with reusable QueryBuilder extensions
- Support first, last, before, after parameters matching Python reference implementation
- Include metadata fields: has_next_page, has_previous_page, start_cursor, end_cursor

**Domain Models and Database Schema**
- Create auth domain models in `/server/internal/domain/auth/model.go` extending core.CoreModel
- Implement Bun struct tags and table aliases following existing account model patterns
- Define proper field types: bytes for credential IDs, arrays for transports, timestamps for expiration
- Include validation methods: IsExpired(), IsValid(), etc.
- Set up proper foreign key relationships to accounts table

**Security and Token Management**
- Implement MD5 hashing for tokens matching Python reference (hashlib.md5)
- Generate cryptographically secure tokens using crypto/rand and secrets.token_hex(32) equivalent
- Handle token expiration with proper time comparisons
- Implement constant-time comparison for sensitive data verification
- Support for WebAuthn credential bytes handling and challenge generation

**Error Handling and Validation**
- Create domain-specific errors: ErrSessionNotFound, ErrTokenExpired, ErrInvalidCredentials, ErrWebAuthnCredentialNotFound
- Implement proper error wrapping with fmt.Errorf and context preservation
- Handle database constraint violations and unique constraint errors

**Database Operations**
- Implement CRUD operations for all repository types using Bun ORM patterns
- Support for batch operations: create_many, delete_many following Python reference
- Handle database connection pooling and context cancellation
- Implement proper query optimization with indexes and foreign keys

**Static Helper Methods**
- Token generation: generate_session_token(), generate_challenge(), generate_recovery_code()
- Hashing functions: hash_session_token(), hash_challenge(), hash_recovery_code()
- Utility functions for WebAuthn: generate_two_factor_secret()
- Reuse existing password hashing functions from account repository

**Testing and Validation**
- Create comprehensive test coverage for all repository methods
- Test pagination edge cases and cursor handling
- Validate token generation uniqueness and expiration handling

**Integration and Dependencies**
- Integrate with existing dependency injection using fx.Lifecycle patterns
- Ensure proper database connection management and graceful shutdown
- Support for environment-specific configuration (development vs production)
- Include proper logging and debugging support

## Visual Design
No visual assets provided for this technical implementation specification.

## Existing Code to Leverage

**Account Repository Pattern** (`/server/internal/domain/account/repo.go`)
- Interface-first design with static methods for token operations
- Repository struct pattern with Bun DB injection
- Error handling patterns with domain-specific errors
- Helper functions for password hashing and token generation

**Core Model Structure** (`/server/internal/domain/core/model.go`)
- Base model with ID, CreatedAt, UpdatedAt fields
- Bun ORM integration patterns
- Consistent database field naming conventions

**Database Infrastructure** (`/server/internal/infrastructure/db/bun.go`)
- Bun DB configuration and connection pooling
- PostgreSQL dialect setup and query hooks
- Lifecycle management for graceful shutdown

**Python Reference Implementation** (https://github.com/hospitaljobsin/hospitaljobsin/blob/staging/server/app/auth/repositories.py)
- Complete repository patterns for all 8 auth components
- Pagination implementation with PaginatedResult and Paginator
- Token generation and hashing patterns using secrets and hashlib
- WebAuthn credential management and challenge handling
- Recovery code generation and validation logic

## Python Reference Implementation

The following Python source code from the reference implementation should be used as a guide for translating patterns to Go. Each repository class shows the exact method signatures, logic patterns, and data flows that need to be implemented in Go using Bun ORM and PostgreSQL.

**Key Translation Notes:**
- MongoDB/Beanie patterns → PostgreSQL/Bun ORM patterns
- `ObjectId` → Go `int` or `uuid.UUID` for primary keys
- `await` → Go error handling patterns
- Pydantic models → Go struct tags with Bun
- Python datetime → Go `time.Time`
- List comprehensions → Go slices with proper iteration
- `secrets.token_hex(32)` → Go crypto/rand with hex encoding
- `hashlib.md5` → Go crypto/md5

### SessionRepo - Python Reference

```python
class SessionRepo:
    async def create(
        self,
        account: Account,
        user_agent: str,
        ip_address: str,
    ) -> str:
        """Create a new session."""
        session_token = self.generate_session_token()
        session = Session(
            account=account,
            token_hash=self.hash_session_token(
                token=session_token,
            ),
            expires_at=datetime.now(UTC)
            + timedelta(
                seconds=USER_SESSION_EXPIRES_IN,
            ),
            user_agent=user_agent,
            ip_address=ip_address,
        )

        await session.save(link_rule=WriteRules.DO_NOTHING)
        return session_token

    @staticmethod
    def generate_session_token() -> str:
        """Generate a new session token."""
        return secrets.token_hex(32)

    @staticmethod
    def hash_session_token(token: str) -> str:
        """Hash session token."""
        return hashlib.md5(token.encode("utf-8")).hexdigest()  # noqa: S324

    async def get(self, token: str, *, fetch_account: bool = False) -> Session | None:
        """Get session by token."""
        if fetch_account:
            return await Session.find_one(
                Session.token_hash == self.hash_session_token(token),
                fetch_links=True,
                nesting_depth=1,
            )
        return await Session.find_one(
            Session.token_hash == self.hash_session_token(token),
        )

    async def get_by_session_account_id(
        self,
        session_id: ObjectId,
        account_id: ObjectId,
        except_session_token: str | None = None,
    ) -> Session | None:
        """Get session by session ID and account ID."""
        if except_session_token:
            return await Session.find_one(
                Session.id == session_id,
                Session.account.id == account_id,
                Session.token_hash != self.hash_session_token(except_session_token),
            )
        return await Session.find_one(
            Session.id == session_id,
            Session.account.id == account_id,
        )

    async def get_all_list(
        self, account_id: ObjectId, except_session_token: str | None = None
    ) -> list[Session]:
        """Get all sessions for an account."""
        return await Session.find(
            Session.account.id == account_id,
            Session.token_hash != self.hash_session_token(except_session_token),
        ).to_list()

    async def get_all_by_account_id(
        self,
        account_id: ObjectId,
        except_session_token: str,
        first: int | None = None,
        last: int | None = None,
        before: str | None = None,
        after: str | None = None,
    ) -> PaginatedResult[Session, ObjectId]:
        """Get all sessions by account ID."""
        paginator: Paginator[Session, ObjectId] = Paginator(
            reverse=True,
            document_cls=Session,
            paginate_by="_id",
            entity_paginate_by="id",
        )

        return await paginator.paginate(
            search_criteria=Session.find(
                Session.account.id == account_id,
                Session.token_hash
                != self.hash_session_token(
                    except_session_token,
                ),
            ),
            first=first,
            last=last,
            before=ObjectId(before) if before else None,
            after=ObjectId(after) if after else None,
        )

    async def delete_by_token(self, token: str) -> None:
        """Delete session by token."""
        await Session.find_one(
            Session.token_hash == self.hash_session_token(token),
        ).delete()

    async def delete(self, session: Session) -> None:
        """Delete session."""
        await session.delete()

    async def delete_many(self, session_ids: list[ObjectId]) -> None:
        """Delete many sessions by ID."""
        await Session.find_many(In(Session.id, session_ids)).delete()

    async def delete_all(self, account_id: ObjectId) -> None:
        """Delete all sessions for an account."""
        await Session.find_many(Session.account.id == account_id).delete()
```

### PasswordResetTokenRepo - Python Reference

```python
class PasswordResetTokenRepo:
    async def create(
        self,
        account: Account,
    ) -> str:
        """Create a new password reset token."""
        token = self.generate_password_reset_token()
        token_hash = self.hash_password_reset_token(token)
        expires_at = datetime.now(UTC) + timedelta(seconds=PASSWORD_RESET_EXPIRES_IN)

        password_reset_token = PasswordResetToken(
            account=account,
            token_hash=token_hash,
            expires_at=expires_at,
        )

        await password_reset_token.save(link_rule=WriteRules.WRITE)
        return token

    @staticmethod
    def generate_password_reset_token() -> str:
        """Generate a new password reset token."""
        return secrets.token_hex(32)

    @staticmethod
    def hash_password_reset_token(token: str) -> str:
        """Hash password reset token."""
        return hashlib.md5(token.encode("utf-8")).hexdigest()  # noqa: S324

    async def get(self, token: str, email: str) -> PasswordResetToken | None:
        """Get password reset token by token."""
        return await PasswordResetToken.find_one(
            PasswordResetToken.token_hash == self.hash_password_reset_token(token),
            PasswordResetToken.account.email == email,
            fetch_links=True,
            nesting_depth=1,
        )

    async def get_by_account(self, account_id: ObjectId) -> PasswordResetToken | None:
        """Get password reset token by account."""
        return await PasswordResetToken.find_one(
            PasswordResetToken.account.id == account_id,
            fetch_links=True,
            nesting_depth=1,
        )

    async def delete(self, token: PasswordResetToken) -> None:
        """Delete password reset token."""
        await token.delete()
```

### WebAuthnCredentialRepo - Python Reference

```python
class WebAuthnCredentialRepo:
    async def create(
        self,
        *,
        account_id: ObjectId,
        credential_id: bytes,
        credential_public_key: bytes,
        sign_count: int,
        device_type: str,
        backed_up: bool,
        transports: list[AuthenticatorTransport] | None = None,
        nickname: str | None = None,
    ) -> WebAuthnCredential:
        web_authn_credential = WebAuthnCredential(
            account=account_id,
            credential_id=credential_id,
            public_key=credential_public_key,
            sign_count=sign_count,
            device_type=device_type,
            transports=transports,
            backed_up=backed_up,
            nickname=nickname,
            last_used_at=datetime.now(UTC),
        )
        await web_authn_credential.save()
        return web_authn_credential

    async def update_sign_count(
        self,
        *,
        web_authn_credential: WebAuthnCredential,
        sign_count: int,
    ) -> None:
        """Update the given WebAuthn credential."""
        web_authn_credential.sign_count = sign_count
        web_authn_credential.last_used_at = datetime.now(UTC)
        await web_authn_credential.save()

    async def get(self, credential_id: bytes) -> WebAuthnCredential | None:
        """Get WebAuthn credential by credential ID."""
        return await WebAuthnCredential.find_one(
            WebAuthnCredential.credential_id == credential_id,
            fetch_links=True,
            nesting_depth=1,
        )

    async def get_by_account_credential_id(
        self, account_id: ObjectId, web_authn_credential_id: ObjectId
    ) -> WebAuthnCredential | None:
        """Get WebAuthn credential by account ID and credential ID."""
        return await WebAuthnCredential.find_one(
            WebAuthnCredential.id == web_authn_credential_id,
            WebAuthnCredential.account.id == account_id,
        )

    async def delete(self, web_authn_credential: WebAuthnCredential) -> None:
        """Delete WebAuthn credential."""
        await web_authn_credential.delete()

    async def get_all_by_account_list(
        self, account_id: ObjectId
    ) -> list[WebAuthnCredential]:
        """Get all webauthn credentials by account ID."""
        return await WebAuthnCredential.find(
            WebAuthnCredential.account.id == account_id,
        ).to_list()

    async def update(
        self, web_authn_credential: WebAuthnCredential, nickname: str
    ) -> WebAuthnCredential:
        """Update WebAuthn credential."""
        web_authn_credential.nickname = nickname
        await web_authn_credential.save()
        return web_authn_credential

    async def get_all_by_account_id(
        self,
        account_id: ObjectId,
        first: int | None = None,
        last: int | None = None,
        before: str | None = None,
        after: str | None = None,
    ) -> PaginatedResult[WebAuthnCredential, ObjectId]:
        """Get all webauthn credentials by account ID."""
        paginator: Paginator[WebAuthnCredential, ObjectId] = Paginator(
            reverse=True,
            document_cls=WebAuthnCredential,
            paginate_by="_id",
            entity_paginate_by="id",
        )

        return await paginator.paginate(
            search_criteria=WebAuthnCredential.find(
                WebAuthnCredential.account.id == account_id,
            ),
            first=first,
            last=last,
            before=ObjectId(before) if before else None,
            after=ObjectId(after) if after else None,
        )
```

### WebAuthnChallengeRepo - Python Reference

```python
class WebAuthnChallengeRepo:
    async def create(
        self, challenge: bytes, generated_account_id: ObjectId
    ) -> WebAuthnChallenge:
        expires_at = datetime.now(UTC) + timedelta(
            seconds=WEBAUTHN_CHALLENGE_EXPIRES_IN
        )
        webauthn_challenge = WebAuthnChallenge(
            challenge=challenge,
            generated_account_id=generated_account_id,
            expires_at=expires_at,
        )
        await webauthn_challenge.save()
        return webauthn_challenge

    async def get(self, challenge: bytes) -> WebAuthnChallenge | None:
        """Get WebAuthn challenge by challenge."""
        return await WebAuthnChallenge.find_one(
            WebAuthnChallenge.challenge == challenge,
        )

    async def delete(self, webauthn_challenge: WebAuthnChallenge) -> None:
        """Delete WebAuthn challenge by ID."""
        await webauthn_challenge.delete()
```

### OauthCredentialRepo - Python Reference

```python
class OauthCredentialRepo:
    async def create(
        self,
        account_id: ObjectId,
        provider: OAuthProvider,
        provider_user_id: str,
    ) -> OAuthCredential:
        """Create a new OAuth credential."""
        oauth_credential = OAuthCredential(
            account=account_id,
            provider=provider,
            provider_user_id=provider_user_id,
        )
        await oauth_credential.save()
        return oauth_credential
```

### TwoFactorAuthenticationChallengeRepo - Python Reference

```python
class TwoFactorAuthenticationChallengeRepo:
    @staticmethod
    def hash_challenge(challenge: str) -> str:
        """Hash the given challenge."""
        return hashlib.md5(challenge.encode("utf-8")).hexdigest()  # noqa: S324

    @staticmethod
    def generate_challenge() -> str:
        """Generate a new challenge."""
        return secrets.token_hex(32)

    @staticmethod
    def generate_two_factor_secret() -> str:
        """Generate a new 2fa secret."""
        return pyotp.random_base32()

    async def create(
        self, *, account: Account, totp_secret: str | None = None
    ) -> tuple[str, TwoFactorAuthenticationChallenge]:
        """Create a new 2FA challenge."""
        challenge = self.generate_challenge()
        expires_at = datetime.now(UTC) + timedelta(
            seconds=TWO_FACTOR_AUTHENTICATION_CHALLENGE_EXPIRES_IN
        )
        two_factor_challenge = TwoFactorAuthenticationChallenge(
            challenge_hash=self.hash_challenge(challenge),
            expires_at=expires_at,
            account=account.id,
            totp_secret=self.generate_two_factor_secret()
            if totp_secret is None
            else totp_secret,
        )
        await two_factor_challenge.save()
        return challenge, two_factor_challenge

    async def get(self, challenge: str) -> TwoFactorAuthenticationChallenge | None:
        """Get 2FA challenge by challenge."""
        return await TwoFactorAuthenticationChallenge.find_one(
            TwoFactorAuthenticationChallenge.challenge_hash
            == self.hash_challenge(challenge),
            fetch_links=True,
            nesting_depth=1,
        )
```

### RecoveryCodeRepo - Python Reference

```python
class RecoveryCodeRepo:
    @staticmethod
    def hash_recovery_code(recovery_code: str) -> str:
        """Hash the given recovery code."""
        return hashlib.md5(recovery_code.encode("utf-8")).hexdigest()  # noqa: S324

    @staticmethod
    def generate_recovery_code() -> str:
        """Generate a new recovery code."""
        charset = string.digits + string.ascii_letters
        return "".join(secrets.choice(charset) for _ in range(8))

    async def create(
        self,
        account_id: ObjectId,
        code: str | None = None,
    ) -> str:
        """Create a recovery code for an account (used in E2E testing)."""
        code = code or self.generate_recovery_code()
        await RecoveryCode(
            code_hash=self.hash_recovery_code(code),
            account=account_id,
        ).save()
        return code

    async def create_many(
        self, account_id: ObjectId, code_count: int = 10
    ) -> list[str]:
        """Create many recovery codes."""
        codes = [self.generate_recovery_code() for _ in range(code_count)]
        recovery_codes = [
            RecoveryCode(
                code_hash=self.hash_recovery_code(code),
                account=account_id,
            )
            for code in codes
        ]
        await RecoveryCode.insert_many(recovery_codes)
        return codes

    async def delete_all(self, account_id: ObjectId) -> None:
        """Delete all recovery codes for an account."""
        await RecoveryCode.find_many(RecoveryCode.account.id == account_id).delete()

    async def delete(self, recovery_code: RecoveryCode) -> None:
        """Delete recovery code."""
        await recovery_code.delete()

    async def get(self, account_id: ObjectId, code: str) -> RecoveryCode | None:
        """Get all recovery codes by account ID."""
        return await RecoveryCode.find_one(
            RecoveryCode.account.id == account_id,
            RecoveryCode.code_hash == self.hash_recovery_code(code),
        )
```

### TemporaryTwoFactorChallengeRepo - Python Reference

```python
class TemporaryTwoFactorChallengeRepo:
    @staticmethod
    def hash_challenge(challenge: str) -> str:
        """Hash the given challenge."""
        return hashlib.md5(challenge.encode("utf-8")).hexdigest()  # noqa: S324

    @staticmethod
    def generate_challenge() -> str:
        """Generate a new challenge."""
        return secrets.token_hex(32)

    async def create(
        self, account_id: ObjectId, password_reset_token_id: ObjectId
    ) -> str:
        """Create a new temporary 2FA challenge."""
        challenge = self.generate_challenge()
        expires_at = datetime.now(UTC) + timedelta(
            seconds=TWO_FACTOR_AUTHENTICATION_CHALLENGE_EXPIRES_IN,
        )
        temporary_two_factor_challenge = TemporaryTwoFactorChallenge(
            challenge_hash=self.hash_challenge(challenge),
            expires_at=expires_at,
            account=account_id,
            password_reset_token=password_reset_token_id,
        )
        await temporary_two_factor_challenge.save()
        return challenge

    async def get(
        self, challenge: str, password_reset_token_id: ObjectId
    ) -> TemporaryTwoFactorChallenge | None:
        """Get temporary 2FA challenge by challenge."""
        return await TemporaryTwoFactorChallenge.find_one(
            TemporaryTwoFactorChallenge.challenge_hash
            == self.hash_challenge(challenge),
            TemporaryTwoFactorChallenge.password_reset_token.id
            == password_reset_token_id,
            fetch_links=True,
            nesting_depth=1,
        )

    async def delete(
        self, temporary_two_factor_challenge: TemporaryTwoFactorChallenge
    ) -> None:
        """Delete temporary 2FA challenge."""
        await temporary_two_factor_challenge.delete()
```

## Out of Scope
- OAuth provider-specific implementations (focus on repository pattern only)
- Frontend UI components or API endpoint implementations
- Authentication service layer or business logic
- Database migration scripts or schema management
- Security audits or penetration testing
- Performance benchmarking and optimization beyond basic patterns
- Internationalization or localization support
- Email/SMS delivery mechanisms for tokens
- Rate limiting or abuse prevention mechanisms
- Audit logging or compliance reporting features