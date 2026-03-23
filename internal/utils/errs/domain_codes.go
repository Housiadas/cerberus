package errs

// Domain-specific error code constants.
// Format: "domain.specific_error".
const (
	// CodeInternal represents an internal server error.
	CodeInternal = "internal.error"
	// CodePanic represents a panic recovery error.
	CodePanic = "internal.panic"

	// CodeUnknown represents an unknown error.
	CodeUnknown = "unknown.error"

	// CodeValidation represents a validation error.
	CodeValidation = "validation.error"

	// CodeRequestInvalid represents an invalid request error.
	CodeRequestInvalid = "request.invalid"

	// CodeUnauthenticated represents an unauthenticated request.
	CodeUnauthenticated = "auth.unauthenticated"
	// CodeInvalidToken represents an invalid token error.
	CodeInvalidToken = "auth.invalid_token"
	// CodeExpiredToken represents an expired token error.
	CodeExpiredToken = "auth.expired_token"
	// CodeUserDisabled represents a disabled user error.
	CodeUserDisabled = "auth.user_disabled"

	// CodeUserNotFound represents a user not found error.
	CodeUserNotFound = "user.not_found"
	// CodeUniqueEmail represents a duplicate email error.
	CodeUniqueEmail = "user.unique_email"
	// CodeAuthFailed represents an authentication failure.
	CodeAuthFailed = "user.auth_failed"

	// CodeRoleNotFound represents a role not found error.
	CodeRoleNotFound = "role.not_found"

	// CodePermissionNotFound represents a permission not found error.
	CodePermissionNotFound = "permission.not_found"
	// CodePermissionDenied represents a permission denied error.
	CodePermissionDenied = "permission.denied"
	// CodePermissionCheckErr represents a permission check error.
	CodePermissionCheckErr = "permission.check_error"

	// CodeAccountNotFound represents an account not found error.
	CodeAccountNotFound = "account.not_found"
)
