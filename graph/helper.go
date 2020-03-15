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
