package widget

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainErrorWithCause creates a new domain error with a cause
func NewDomainErrorWithCause(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ErrorCode represents domain error codes
type ErrorCode string

const (
	// Widget errors
	ErrWidgetNotFound       ErrorCode = "WIDGET_NOT_FOUND"
	ErrWidgetAlreadyExists  ErrorCode = "WIDGET_ALREADY_EXISTS"
	ErrInvalidName          ErrorCode = "INVALID_NAME"
	ErrInvalidTemplateType  ErrorCode = "INVALID_TEMPLATE_TYPE"
	ErrInvalidDataSource    ErrorCode = "INVALID_DATA_SOURCE"
	ErrInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
	ErrMissingConfiguration ErrorCode = "MISSING_CONFIGURATION"
	ErrInvalidURL           ErrorCode = "INVALID_URL"
	ErrMissingRequiredField ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrInvalidFieldMapping  ErrorCode = "INVALID_FIELD_MAPPING"

	// Repository errors
	ErrRepositoryFailure  ErrorCode = "REPOSITORY_FAILURE"
	ErrDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrTransactionFailure ErrorCode = "TRANSACTION_FAILURE"

	// Service errors
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrValidationFailure  ErrorCode = "VALIDATION_FAILURE"
	ErrPermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
)

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrWidgetNotFound
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		switch domainErr.Code {
		case ErrInvalidName, ErrInvalidTemplateType, ErrInvalidDataSource,
			ErrInvalidConfiguration, ErrMissingConfiguration, ErrInvalidURL,
			ErrMissingRequiredField, ErrInvalidFieldMapping, ErrValidationFailure:
			return true
		}
	}
	return false
}

// IsRepositoryError checks if the error is a repository error
func IsRepositoryError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		switch domainErr.Code {
		case ErrRepositoryFailure, ErrDatabaseConnection, ErrTransactionFailure:
			return true
		}
	}
	return false
}
