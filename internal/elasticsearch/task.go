package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	esv7 "github.com/elastic/go-elasticsearch/v7"
	esv7api "github.com/elastic/go-elasticsearch/v7/esapi"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

const otelName = "github.com/MarioCarrion/todo-api-microservice-example/internal/elasticsearch"

// Task represents the repository used for interacting with Task records.
type Task struct {
	client *esv7.Client
	index  string
}

//nolint:tagliatelle
type indexedTask struct {
	// XXX: `SubTasks` and `Categories` will be added in future episodes
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Priority    internal.Priority `json:"priority"`
	IsDone      bool              `json:"is_done"`
	DateStart   int64             `json:"date_start"`
	DateDue     int64             `json:"date_due"`
}

// NewTask instantiates the Task repository.
func NewTask(client *esv7.Client) *Task {
	return &Task{
		client: client,
		index:  "tasks",
	}
}

// Index creates or updates a task in an index.
func (t *Task) Index(ctx context.Context, task internal.Task) error {
	defer newOTELSpan(ctx, "Task.Index").End()

	//-

	body := indexedTask{
		ID:          task.ID,
		Description: task.Description,
		Priority:    task.Priority,
		IsDone:      task.IsDone,
		DateStart:   task.Dates.Start.UnixNano(),
		DateDue:     task.Dates.Due.UnixNano(),
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.NewEncoder.Encode")
	}

	req := esv7api.IndexRequest{
		Index:      t.index,
		Body:       &buf,
		DocumentID: task.ID,
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, t.client)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "IndexRequest.Do")
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return internal.NewErrorf(internal.ErrorCodeUnknown, "IndexRequest.Do %d", resp.StatusCode)
	}

	io.Copy(io.Discard, resp.Body) //nolint: errcheck

	return nil
}

// Delete removes a task from the index.
func (t *Task) Delete(ctx context.Context, id string) error {
	defer newOTELSpan(ctx, "Task.Delete").End()

	//-

	req := esv7api.DeleteRequest{
		Index:      t.index,
		DocumentID: id,
	}

	resp, err := req.Do(ctx, t.client)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "DeleteRequest.Do")
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return internal.NewErrorf(internal.ErrorCodeUnknown, "DeleteRequest.Do %d", resp.StatusCode)
	}

	io.Copy(io.Discard, resp.Body) //nolint: errcheck

	return nil
}

// Search returns tasks matching a query.
//
//nolint:funlen,cyclop
func (t *Task) Search(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error) {
	defer newOTELSpan(ctx, "Task.Search").End()

	//-

	if args.IsZero() {
		return internal.SearchResults{}, nil
	}

	should := make([]interface{}, 0, 3)

	if args.Description != nil {
		should = append(should, map[string]interface{}{
			"match": map[string]interface{}{
				"description": *args.Description,
			},
		})
	}

	if args.Priority != nil {
		should = append(should, map[string]interface{}{
			"match": map[string]interface{}{
				"priority": *args.Priority,
			},
		})
	}

	if args.IsDone != nil {
		should = append(should, map[string]interface{}{
			"match": map[string]interface{}{
				"is_done": *args.IsDone,
			},
		})
	}

	var query map[string]interface{}

	if len(should) > 1 {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"should": should,
				},
			},
		}
	} else {
		query = map[string]interface{}{
			"query": should[0],
		}
	}

	query["sort"] = []interface{}{
		"_score",
		map[string]interface{}{"id": "asc"},
	}

	query["from"] = args.From
	query["size"] = args.Size

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.NewEncoder.Encode")
	}

	req := esv7api.SearchRequest{
		Index: []string{t.index},
		Body:  &buf,
	}

	resp, err := req.Do(ctx, t.client)
	if err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "SearchRequest.Do")
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return internal.SearchResults{}, internal.NewErrorf(internal.ErrorCodeUnknown, "SearchRequest.Do %d", resp.StatusCode)
	}

	//nolint: tagliatelle
	var hits struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source indexedTask `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&hits); err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.NewDecoder.Decode")
	}

	res := make([]internal.Task, len(hits.Hits.Hits))

	for i, hit := range hits.Hits.Hits {
		res[i].ID = hit.Source.ID
		res[i].Description = hit.Source.Description
		res[i].Priority = hit.Source.Priority
		res[i].Dates.Due = time.Unix(0, hit.Source.DateDue).UTC()
		res[i].Dates.Start = time.Unix(0, hit.Source.DateStart).UTC()
	}

	return internal.SearchResults{
		Tasks: res,
		Total: hits.Hits.Total.Value,
	}, nil
}

//-

func newOTELSpan(ctx context.Context, name string) trace.Span {
	_, span := otel.Tracer(otelName).Start(ctx, name)

	span.SetAttributes(semconv.DBSystemElasticsearch)

	return span
}
