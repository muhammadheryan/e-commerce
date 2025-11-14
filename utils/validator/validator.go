package validatorx

import (
	"sync"

	gpvalidator "github.com/go-playground/validator/v10"
)

var (
	v   *gpvalidator.Validate
	mut sync.Mutex
)

// Init initializes the validator singleton (idempotent)
func Init() {
	mut.Lock()
	defer mut.Unlock()
	if v != nil {
		return
	}
	v = gpvalidator.New()
}

// ValidateStruct validates a struct using go-playground/validator
func ValidateStruct(s interface{}) error {
	if v == nil {
		Init()
	}
	return v.Struct(s)
}
