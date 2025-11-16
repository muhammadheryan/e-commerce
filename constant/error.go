package constant

import "net/http"

type ErrorType int

const (
	Successful ErrorType = iota
	ErrInternal
	ErrNotFound
	ErrInvalidRequest
	ErrUnauthorize
	ErrCredentialExists
	ErrInvalidPassword
	ErrInsufficientStock
	ErrInvalidOrderStatus
)

var ErrorTypeMessage = map[ErrorType]string{
	Successful:            "success",
	ErrInternal:           "error internal",
	ErrNotFound:           "data not found",
	ErrInvalidRequest:     "invalid request",
	ErrUnauthorize:        "unauthorize request",
	ErrCredentialExists:   "email or phone already exists",
	ErrInvalidPassword:    "password invalid",
	ErrInsufficientStock:  "insufficient stock",
	ErrInvalidOrderStatus: "invalid order status",
}

var ErrorTypeHTTPCode = map[ErrorType]int{
	Successful:            http.StatusOK,
	ErrInternal:           http.StatusInternalServerError,
	ErrNotFound:           http.StatusBadRequest,
	ErrInvalidRequest:     http.StatusBadRequest,
	ErrUnauthorize:        http.StatusUnauthorized,
	ErrCredentialExists:   http.StatusBadRequest,
	ErrInvalidPassword:    http.StatusBadRequest,
	ErrInsufficientStock:  http.StatusBadRequest,
	ErrInvalidOrderStatus: http.StatusBadRequest,
}

var ErrorTypeCode = map[ErrorType]string{
	Successful:            "0000",
	ErrInternal:           "0001",
	ErrNotFound:           "0002",
	ErrInvalidRequest:     "0003",
	ErrUnauthorize:        "0004",
	ErrCredentialExists:   "0005",
	ErrInvalidPassword:    "0006",
	ErrInsufficientStock:  "0007",
	ErrInvalidOrderStatus: "0008",
}
