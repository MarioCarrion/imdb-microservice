package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/rest"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/rest/resttesting"
)

func TestTasks_Delete(t *testing.T) {
	t.Parallel()

	// XXX: Test "serviceArgs"

	type output struct {
		expectedStatus int
		expected       interface{}
		target         interface{}
	}

	tests := []struct {
		name   string
		setup  func(*resttesting.FakeTaskService)
		output output
	}{
		{
			"OK: 200",
			func(_ *resttesting.FakeTaskService) {},
			output{
				http.StatusOK,
				&struct{}{},
				&struct{}{},
			},
		},
		{
			"ERR: 404",
			func(s *resttesting.FakeTaskService) {
				s.DeleteReturns(internal.NewErrorf(internal.ErrorCodeNotFound, "not found"))
			},
			output{
				http.StatusNotFound,
				&struct{}{},
				&struct{}{},
			},
		},
		{
			"ERR: 500",
			func(s *resttesting.FakeTaskService) {
				s.DeleteReturns(errors.New("service failed"))
			},
			output{
				http.StatusInternalServerError,
				&struct{}{},
				&struct{}{},
			},
		},
	}

	//-

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newRouter()
			svc := &resttesting.FakeTaskService{}
			tt.setup(svc)

			rest.NewTaskHandler(svc).Register(router)

			//-

			res := doRequest(router,
				httptest.NewRequest(http.MethodDelete, "/tasks/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", nil))

			//-

			assertResponse(t, res, test{tt.output.expected, tt.output.target})

			if tt.output.expectedStatus != res.StatusCode {
				t.Fatalf("expected code %d, actual %d", tt.output.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestTasks_Post(t *testing.T) {
	t.Parallel()

	// XXX: Test "serviceArgs"

	type output struct {
		expectedStatus int
		expected       interface{}
		target         interface{}
	}

	tests := []struct {
		name   string
		setup  func(*resttesting.FakeTaskService)
		input  []byte
		output output
	}{
		{
			"OK: 201",
			func(s *resttesting.FakeTaskService) {
				s.CreateReturns(
					internal.Task{
						ID:          "1-2-3",
						Description: "new task",
						Priority:    internal.PriorityHigh,
					},
					nil)
			},
			func() []byte {
				b, _ := json.Marshal(&rest.CreateTasksRequest{
					Description: "new task",
					Priority:    "high",
				})

				return b
			}(),
			output{
				http.StatusCreated,
				&rest.CreateTasksResponse{
					Task: rest.Task{
						ID:          "1-2-3",
						Description: "new task",
						Priority:    "high",
					},
				},
				&rest.CreateTasksResponse{},
			},
		},
		{
			"ERR: 400",
			func(*resttesting.FakeTaskService) {},
			[]byte(`{"invalid":"json`),
			output{
				http.StatusBadRequest,
				&rest.ErrorResponse{
					Error: "invalid request",
				},
				&rest.ErrorResponse{},
			},
		},
		{
			"ERR: 500",
			func(s *resttesting.FakeTaskService) {
				s.CreateReturns(internal.Task{},
					errors.New("service error"))
			},
			[]byte(`{}`),
			output{
				http.StatusInternalServerError,
				&rest.ErrorResponse{
					Error: "internal error",
				},
				&rest.ErrorResponse{},
			},
		},
	}

	//-

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newRouter()
			svc := &resttesting.FakeTaskService{}
			tt.setup(svc)

			rest.NewTaskHandler(svc).Register(router)

			//-

			res := doRequest(router,
				httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(tt.input)))

			//-

			assertResponse(t, res, test{tt.output.expected, tt.output.target})

			if tt.output.expectedStatus != res.StatusCode {
				t.Fatalf("expected code %d, actual %d", tt.output.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestTasks_Read(t *testing.T) {
	t.Parallel()

	// XXX: Test "serviceArgs"

	type output struct {
		expectedStatus int
		expected       interface{}
		target         interface{}
	}

	tests := []struct {
		name   string
		setup  func(*resttesting.FakeTaskService)
		output output
	}{
		{
			"OK: 200",
			func(s *resttesting.FakeTaskService) {
				s.TaskReturns(
					internal.Task{
						ID:          "a-b-c",
						Description: "existing task",
						IsDone:      true,
					},
					nil)
			},
			output{
				http.StatusOK,
				&rest.ReadTasksResponse{
					Task: rest.Task{
						ID:          "a-b-c",
						Description: "existing task",
						Priority:    "none",
						IsDone:      true,
					},
				},
				&rest.ReadTasksResponse{},
			},
		},
		{
			"OK: 200",
			func(s *resttesting.FakeTaskService) {
				s.TaskReturns(internal.Task{},
					internal.NewErrorf(internal.ErrorCodeNotFound, "not found"))
			},
			output{
				http.StatusNotFound,
				&rest.ErrorResponse{
					Error: "find failed",
				},
				&rest.ErrorResponse{},
			},
		},
		{
			"ERR: 500",
			func(s *resttesting.FakeTaskService) {
				s.TaskReturns(internal.Task{},
					errors.New("service error"))
			},
			output{
				http.StatusInternalServerError,
				&rest.ErrorResponse{
					Error: "internal error",
				},
				&rest.ErrorResponse{},
			},
		},
	}

	//-

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newRouter()
			svc := &resttesting.FakeTaskService{}
			tt.setup(svc)

			rest.NewTaskHandler(svc).Register(router)

			//-

			res := doRequest(router,
				httptest.NewRequest(http.MethodGet, "/tasks/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", nil))

			//-

			assertResponse(t, res, test{tt.output.expected, tt.output.target})

			if tt.output.expectedStatus != res.StatusCode {
				t.Fatalf("expected code %d, actual %d", tt.output.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestTasks_Update(t *testing.T) {
	t.Parallel()

	// XXX: Test "serviceArgs"

	type output struct {
		expectedStatus int
		expected       interface{}
		target         interface{}
	}

	tests := []struct {
		name   string
		setup  func(*resttesting.FakeTaskService)
		input  []byte
		output output
	}{
		{
			"OK: 200",
			func(_ *resttesting.FakeTaskService) {},
			func() []byte {
				b, _ := json.Marshal(&rest.UpdateTasksRequest{
					Description: "update task",
					Priority:    "low",
				})

				return b
			}(),
			output{
				http.StatusOK,
				&struct{}{},
				&struct{}{},
			},
		},
		{
			"ERR: 400",
			func(*resttesting.FakeTaskService) {},
			[]byte(`{"invalid":"json`),
			output{
				http.StatusBadRequest,
				&rest.ErrorResponse{
					Error: "invalid request",
				},
				&rest.ErrorResponse{},
			},
		},
		{
			"ERR: 404",
			func(s *resttesting.FakeTaskService) {
				s.UpdateReturns(internal.NewErrorf(internal.ErrorCodeNotFound, "not found"))
			},
			func() []byte {
				b, _ := json.Marshal(&rest.UpdateTasksRequest{
					Description: "update task",
					Priority:    "low",
				})

				return b
			}(),
			output{
				http.StatusNotFound,
				&struct{}{},
				&struct{}{},
			},
		},
		{
			"ERR: 500",
			func(s *resttesting.FakeTaskService) {
				s.UpdateReturns(errors.New("service error"))
			},
			[]byte(`{}`),
			output{
				http.StatusInternalServerError,
				&rest.ErrorResponse{
					Error: "internal error",
				},
				&rest.ErrorResponse{},
			},
		},
	}

	//-

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newRouter()
			svc := &resttesting.FakeTaskService{}
			tt.setup(svc)

			rest.NewTaskHandler(svc).Register(router)

			//-

			res := doRequest(router,
				httptest.NewRequest(http.MethodPut, "/tasks/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", bytes.NewReader(tt.input)))

			//-

			assertResponse(t, res, test{tt.output.expected, tt.output.target})

			if tt.output.expectedStatus != res.StatusCode {
				t.Fatalf("expected code %d, actual %d", tt.output.expectedStatus, res.StatusCode)
			}
		})
	}
}

type test struct {
	expected interface{}
	target   interface{}
}

func doRequest(router *chi.Mux, req *http.Request) *http.Response {
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	return rr.Result()
}

func assertResponse(t *testing.T, res *http.Response, test test) {
	t.Helper()

	if err := json.NewDecoder(res.Body).Decode(test.target); err != nil {
		t.Fatalf("couldn't decode %s", err)
	}
	defer res.Body.Close()

	if !cmp.Equal(test.expected, test.target, cmpopts.IgnoreUnexported(time.Time{})) {
		t.Fatalf("expected results don't match: %s", cmp.Diff(test.expected, test.target, cmpopts.IgnoreUnexported(time.Time{})))
	}
}

func newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))

	return r
}
