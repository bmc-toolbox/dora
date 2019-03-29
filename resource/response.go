package resource

import "errors"

// The Response struct implements api2go.Responder
type Response struct {
	Res  interface{}
	Code int
}

var (
	// ErrPageSizeAndNumber is returned when page[number] and page[size] are sent on the http request
	ErrPageSizeAndNumber = errors.New("filters page[number] and page[size] are not supported, please stick to page[offset] and page[limit]")
)

// Metadata returns additional meta data
func (r Response) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"author":      "PSM Crew",
		"service":     "Dora",
		"license":     "APACHE2",
		"license-url": "https://www.apache.org/licenses/LICENSE-2.0",
	}
}

// Result returns the actual payload
func (r Response) Result() interface{} {
	return r.Res
}

// StatusCode sets the return status code
func (r Response) StatusCode() int {
	return r.Code
}
