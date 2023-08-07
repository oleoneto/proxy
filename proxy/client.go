package proxy

import (
	"net/http"
)

var HTTPClient http.Client

func InitializeHTTPClient() {
	once.Do(func() {
		HTTPClient = http.Client{}
	})
}
