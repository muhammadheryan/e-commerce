package transport

import (
	"encoding/json"
	"net/http"

	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/utils/errors"
)

type body struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	customError, ok := err.(errors.CustomError)
	if !ok {
		customError = errors.SetCustomError(constant.ErrInternal)
	}

	data := body{
		Code:    customError.ErrorCode(),
		Message: customError.Error(),
	}
	writeJson(w, customError.ErrorHTTPCode(), data)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJson(w, http.StatusOK, body{
		Code:    constant.ErrorTypeCode[constant.Successful],
		Message: constant.ErrorTypeMessage[constant.Successful],
		Data:    data,
	})
}
