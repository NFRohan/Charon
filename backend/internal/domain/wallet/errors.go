package wallet

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
	ErrWalletAccountNotFound = &Error{
		code:    "WALLET_ACCOUNT_NOT_FOUND",
		message: "Wallet account not found.",
	}
)

func ErrorCode(err error) string {
	var walletErr *Error
	if errors.As(err, &walletErr) {
		return walletErr.Code()
	}

	return "INTERNAL_ERROR"
}

func ErrorMessage(err error) string {
	var walletErr *Error
	if errors.As(err, &walletErr) {
		return walletErr.Message()
	}

	return "An unexpected error occurred."
}
