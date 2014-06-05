package jantar

import (
	"net/http"
)

// StatusHandler is a map containing a http.HandlerFunc for each client side http status code. This allows
// developer to add their own custom http.HandlerFunc for given status codes.
var StatusHandler map[int]func(http.ResponseWriter, *http.Request)

func init() {
	statusResponse := map[int]string{
		http.StatusBadRequest:                   "400 bad request",
		http.StatusUnauthorized:                 "401 unauthorized",
		http.StatusPaymentRequired:              "402 payment required",
		http.StatusForbidden:                    "403 forbidden",
		http.StatusNotFound:                     "404 not found",
		http.StatusMethodNotAllowed:             "405 method not allowed",
		http.StatusNotAcceptable:                "406 not acceptable",
		http.StatusProxyAuthRequired:            "407 proxy auth required",
		http.StatusRequestTimeout:               "408 request timeout",
		http.StatusConflict:                     "409 conflict",
		http.StatusGone:                         "410 gone",
		http.StatusLengthRequired:               "411 length required",
		http.StatusPreconditionFailed:           "412 precondition failed",
		http.StatusRequestEntityTooLarge:        "413 request entity too large",
		http.StatusRequestURITooLong:            "414 request uri too long",
		http.StatusUnsupportedMediaType:         "415 unsupported media type",
		http.StatusRequestedRangeNotSatisfiable: "416 requested range not satisfiable",
		http.StatusExpectationFailed:            "417 expectation failed",
		http.StatusTeapot:                       "418 teapot",
	}

	StatusHandler = make(map[int]func(http.ResponseWriter, *http.Request))
	for status, response := range statusResponse {
		response := response

		StatusHandler[status] = func(respw http.ResponseWriter, req *http.Request) {
			Log.Warning(response)
			respw.Write([]byte(response))
		}
	}
}

// ErrorHandler returns a http.HandlerFunc for a given http status code or nil if no handler can be found for that code.
// Developer can add their own handler by changing the StatusHandler map.
// Note that only 4xx codes are handled by default as 1xx, 2xx and 3xx are no error codes and 5xx should not be catchable.
func ErrorHandler(status int) func(http.ResponseWriter, *http.Request) {
	if handler, ok := StatusHandler[status]; ok {
		return handler
	}
	return nil
}
