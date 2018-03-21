package service_error

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrInvalidResponseFormat is an error returned by unexpected external service's response.
var ErrInvalidResponseFormat = errors.New("Invalid response format from API")

// ServiceError uses to separate critical and non-critical errors which returns in external service response.
// For this type of error we shouldn't use 500 error counter for librato
// easyjson:json
type ServiceError struct {
	Code    int            `json:"error_code"`
	Message string         `json:"error_message"`
	ErrorV2 serviceErrorV2 `json:"error"`
	Success bool           `json:"success"`
}

func (err ServiceError) Error() string {
	return err.Message
}

type serviceErrorV2 struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func AddServiceName(err error, serviceName string) error {
	if e, ok := err.(ServiceError); ok {
		e.Message = AddServiceNameText(e.Message, serviceName)
		return e
	}

	return errors.New(AddServiceNameText(err.Error(), serviceName))
}

func AddServiceNameText(errText, serviceName string) string {
	return fmt.Sprintf("%s: %s", serviceName, errText)
}

func AddServiceNameToErrorStruct(errorStruct interface{}, serviceName string) {
	errorStructReflect := reflect.ValueOf(errorStruct).Elem()
	for i := 0; i < errorStructReflect.NumField(); i++ {
		errInterface := errorStructReflect.Field(i).Interface()
		if err, ok := errInterface.(error); ok {
			err = AddServiceName(err, serviceName)
			errorStructReflect.Field(i).Set(reflect.ValueOf(err))
		}
	}
}
