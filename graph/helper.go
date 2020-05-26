package graph

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// NewGQLError _
func NewGQLError(vars ...string) error {
	if len(vars) < 1 {
		panic("must have 1 variable at least")
	}

	err := &gqlerror.Error{
		Message: vars[0],
	}

	if len(vars) > 1 {
		err.Extensions = map[string]interface{}{
			"originErr": vars[1],
		}
	}

	return err
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
