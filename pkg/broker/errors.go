package broker

import (
	"net/http"

	"github.com/pmorie/go-open-service-broker-client/v2"
)

var (
	asyncRequiredError = newAsyncRequiredError()
	concurrencyError = newConcurrencyError()
)

func newAsyncRequiredError() v2.HTTPStatusCodeError {
	return v2.HTTPStatusCodeError{
		StatusCode:   http.StatusUnprocessableEntity,
		ErrorMessage: &[]string{v2.AsyncErrorMessage}[0],
		Description:  &[]string{v2.AsyncErrorDescription}[0],
	}
}

func newConcurrencyError() v2.HTTPStatusCodeError {
	return v2.HTTPStatusCodeError{
		StatusCode:   http.StatusUnprocessableEntity,
		ErrorMessage: &[]string{v2.ConcurrencyErrorMessage}[0],
		Description:  &[]string{v2.ConcurrencyErrorDescription}[0],
	}
}
