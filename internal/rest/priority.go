package rest

import (
	"encoding/json"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

// Priority indicates how important a Task is.
type Priority string

const (
	priorityNone   Priority = "none"
	priorityLow    Priority = "low"
	priorityMedium Priority = "medium"
	priorityHigh   Priority = "high"
)

// NewPriority converts the received domain type to a rest type, when the argument is unknown "none" is used.
func NewPriority(p internal.Priority) Priority {
	switch p {
	case internal.PriorityNone:
		return priorityNone
	case internal.PriorityLow:
		return priorityLow
	case internal.PriorityMedium:
		return priorityMedium
	case internal.PriorityHigh:
		return priorityHigh
	}

	return priorityNone
}

// Convert returns the domain type defining the internal representation, when priority is unknown "none" is
// used.
func (p *Priority) Convert() internal.Priority {
	switch *p {
	case priorityNone:
		return internal.PriorityNone
	case priorityLow:
		return internal.PriorityLow
	case priorityMedium:
		return internal.PriorityMedium
	case priorityHigh:
		return internal.PriorityHigh
	}

	return internal.PriorityNone
}

// Validate ...
func (p *Priority) Validate() error {
	switch *p {
	case priorityNone, priorityLow, priorityMedium, priorityHigh:
		return nil
	}

	return internal.NewErrorf(internal.ErrorCodeInvalidArgument, "unknown value")
}

// MarshalJSON ...
func (p *Priority) MarshalJSON() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "Validate")
	}

	b, err := json.Marshal(string(*p))
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.Marshal")
	}

	return b, nil
}

// UnmarshalJSON ...
func (p *Priority) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json.Unmarshal")
	}

	conv := Priority(str)

	if err := conv.Validate(); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "Validate")
	}

	*p = Priority(str)

	return nil
}
