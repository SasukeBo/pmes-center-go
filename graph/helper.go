package graph

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// NewGQLError _
func NewGQLError(message, originErr string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"originErr": originErr,
		},
	}
}

func stringP(s string) *string {
	return &s
}
func boolP(b bool) *bool {
	return &b
}
func intP(i int) *int {
	return &i
}
