package envvar_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar/envvartesting"
)

func TestConfiguration_Get(t *testing.T) {
	type output struct {
		val     string
		withErr bool
	}

	tests := []struct {
		name   string
		setup  func(t *testing.T, p *envvartesting.FakeProvider)
		input  string
		output output
		arg    string
	}{
		{
			"OK: no secret",
			func(t *testing.T, _ *envvartesting.FakeProvider) {
				t.Setenv("ENVVAR_OK", "value")
			},
			"ENVVAR_OK",
			output{
				val: "value",
			},
			"",
		},
		{
			"OK: secure",
			func(t *testing.T, p *envvartesting.FakeProvider) {
				t.Setenv("ENVVAR_OK1_SECURE", "/secret/value")

				p.GetReturns("provider value", nil)
			},
			"ENVVAR_OK1",
			output{
				val: "provider value",
			},
			"/secret/value",
		},
		{
			"ERR: provider failed",
			func(t *testing.T, p *envvartesting.FakeProvider) {
				t.Setenv("ENVVAR_ERR_SECURE", "/failed")

				p.GetReturns("", errors.New("failed"))
			},
			"ENVVAR_ERR",
			output{
				withErr: true,
			},
			"/failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := envvartesting.FakeProvider{}

			tt.setup(t, &provider)

			actual, actualErr := envvar.New(&provider).Get(tt.input)

			if (actualErr != nil) != tt.output.withErr {
				t.Fatalf("expected error %t, got %s", tt.output.withErr, actualErr)
			}

			if !cmp.Equal(tt.output.val, actual) {
				t.Fatalf("expected result does not match: %s", cmp.Diff(tt.output.val, actual))
			}

			//- Provider Args

			if provider.GetCallCount() > 0 {
				if arg := provider.GetArgsForCall(0); arg != tt.arg {
					t.Fatalf("expected arg %s, got %s", arg, tt.arg)
				}
			}
		})
	}
}
