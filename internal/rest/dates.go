package rest

import (
	"time"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

// Dates indicates a point in time where a task starts or completes, dates are not enforced on Tasks.
type Dates struct {
	Start time.Time `json:"start"`
	Due   time.Time `json:"due"`
}

// NewDates ...
func NewDates(d internal.Dates) Dates {
	return Dates{
		Start: d.Start,
		Due:   d.Due,
	}
}

// Convert returns the domain type defining the internal representation.
func (d Dates) Convert() internal.Dates {
	return internal.Dates{
		Start: d.Start,
		Due:   d.Due,
	}
}
