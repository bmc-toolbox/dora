package resolver

import (
	"fmt"
	"net/http"
)

type RequestURL struct {
	r    http.Request
	Port int
}

// SetRequest to implement `RequestAwareResolverInterface`
func (m *RequestURL) SetRequest(r http.Request) {
	m.r = r
}

// GetBaseURL implements `URLResolver` interface
func (m RequestURL) GetBaseURL() string {
	return fmt.Sprintf("https://localhost:%d", m.Port)
}
