package validatorx

import (
	"sync"

	gpvalidator "github.com/go-playground/validator/v10"
)

var (
	v   *gpvalidator.Validate
	mut sync.Mutex
)

func Init() {
	mut.Lock()
	defer mut.Unlock()
	if v != nil {
		return
	}
	v = gpvalidator.New()
}

func ValidateStruct(s interface{}) error {
	if v == nil {
		Init()
	}
	return v.Struct(s)
}
