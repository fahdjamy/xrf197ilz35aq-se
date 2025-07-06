package processor

import "fmt"

type ExternalError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *ExternalError) Error() string {
	return fmt.Sprintf("processing error: %s", e.Message)
}

type Processors struct {
	UserProcessor UserProcessor
}
