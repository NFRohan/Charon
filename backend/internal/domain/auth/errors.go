package auth

import "errors"

type Error struct {
	code    string
	message string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Code() string {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

var (
	ErrInvalidCredentials = &Error{
		code:    "INVALID_CREDENTIALS",
		message: "Invalid credentials.",
	}
	ErrAccountDisabled = &Error{
		code:    "ACCOUNT_DISABLED",
		message: "This account is not allowed to sign in.",
	}
	ErrRefreshTokenInvalid = &Error{
		code:    "REFRESH_TOKEN_INVALID",
		message: "Refresh token is invalid.",
	}
	ErrRefreshTokenExpired = &Error{
		code:    "REFRESH_TOKEN_EXPIRED",
		message: "Refresh token has expired.",
	}
	ErrAccessTokenInvalid = &Error{
		code:    "ACCESS_TOKEN_INVALID",
		message: "Access token is invalid.",
	}
	ErrAccessTokenExpired = &Error{
		code:    "ACCESS_TOKEN_EXPIRED",
		message: "Access token has expired.",
	}
	ErrAuthorizationRequired = &Error{
		code:    "AUTHORIZATION_REQUIRED",
		message: "Authorization is required.",
	}
	ErrInsufficientRole = &Error{
		code:    "INSUFFICIENT_ROLE",
		message: "You do not have permission to access this resource.",
	}
)

func ErrorCode(err error) string {
	var authErr *Error
	if errors.As(err, &authErr) {
		return authErr.Code()
	}

	return "INTERNAL_ERROR"
}

func ErrorMessage(err error) string {
	var authErr *Error
	if errors.As(err, &authErr) {
		return authErr.Message()
	}

	return "An unexpected error occurred."
}
