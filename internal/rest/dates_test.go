package rest_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/rest"
)

func TestDates_Marshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  rest.Dates
		output []byte
	}{
		{
			"OK",
			rest.Dates{
				Start: time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
			},
			[]byte(`{"start":"2009-11-10T23:00:00Z","due":"0001-01-01T00:00:00Z"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualRes, actualErr := json.Marshal(&tt.input)
			if actualErr != nil {
				t.Fatalf("expected no error, got %s", actualErr)
			}

			if !cmp.Equal(tt.output, actualRes) {
				t.Fatalf("expected output do not match\n%s", cmp.Diff(tt.output, actualRes))
			}
		})
	}
}

func TestNewDates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  internal.Dates
		output rest.Dates
	}{
		{
			"OK",
			internal.Dates{
				Start: time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
				Due:   time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC).Add(time.Hour),
			},
			rest.Dates{
				Start: time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
				Due:   time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC).Add(time.Hour),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualRes := rest.NewDates(tt.input)
			if !cmp.Equal(tt.output, actualRes, cmpopts.IgnoreUnexported(time.Time{})) {
				t.Fatalf("expected output do not match\n%s", cmp.Diff(tt.output, actualRes, cmpopts.IgnoreUnexported(time.Time{})))
			}
		})
	}
}

func TestDates_Convert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  rest.Dates
		output internal.Dates
	}{
		{
			"OK",
			rest.Dates{
				Start: time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
				Due:   time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC).Add(time.Hour),
			},
			internal.Dates{
				Start: time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
				Due:   time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC).Add(time.Hour),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualRes := tt.input.Convert()
			if !cmp.Equal(tt.output, actualRes, cmpopts.IgnoreUnexported(time.Time{})) {
				t.Fatalf("expected output do not match\n%s", cmp.Diff(tt.output, actualRes, cmpopts.IgnoreUnexported(time.Time{})))
			}
		})
	}
}
