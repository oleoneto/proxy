package proxy

import (
	"net/http"
	"sync"
)

var HTTPClient http.Client

var clientOnce sync.Once

func InitializeHTTPClient() {
	clientOnce.Do(func() {
		HTTPClient = http.Client{}
	})
}
