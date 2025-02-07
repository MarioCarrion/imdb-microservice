package postgresql

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/postgresql/db"
)

// Task represents the repository used for interacting with Task records.
type Task struct {
	q *db.Queries
}

// NewTask instantiates the Task repository.
func NewTask(d db.DBTX) *Task {
	return &Task{
		q: db.New(d),
	}
}

// Create inserts a new task record.
func (t *Task) Create(ctx context.Context, params internal.CreateParams) (internal.Task, error) {
	defer newOTELSpan(ctx, "Task.Create").End()

	//-

	// XXX: `ID` and `IsDone` make no sense when creating new records, that's why those are ignored.
	// XXX: We are intentionally NOT SUPPORTING `SubTasks` and `Categories` JUST YET.

	newID, err := t.q.InsertTask(ctx, db.InsertTaskParams{
		Description: params.Description,
		Priority:    newPriority(params.Priority),
		StartDate:   newTimestamp(params.Dates.Start),
		DueDate:     newTimestamp(params.Dates.Due),
	})
	if err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "insert task")
	}

	return internal.Task{
		ID:          newID.String(),
		Description: params.Description,
		Priority:    params.Priority,
		Dates:       params.Dates,
	}, nil
}

// Delete deletes the existing record matching the id.
func (t *Task) Delete(ctx context.Context, id string) error {
	defer newOTELSpan(ctx, "Task.Delete").End()

	//-

	val, err := uuid.Parse(id)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid uuid")
	}

	_, err = t.q.DeleteTask(ctx, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return internal.WrapErrorf(err, internal.ErrorCodeNotFound, "task not found")
		}

		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "delete task")
	}

	return nil
}

// Find returns the requested task by searching its id.
func (t *Task) Find(ctx context.Context, id string) (internal.Task, error) {
	defer newOTELSpan(ctx, "Task.Find").End()

	//-

	val, err := uuid.Parse(id)
	if err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid uuid")
	}

	res, err := t.q.SelectTask(ctx, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return internal.Task{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "task not found")
		}

		return internal.Task{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "select task")
	}

	priority, err := convertPriority(res.Priority)
	if err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "convert priority")
	}

	return internal.Task{
		ID:          res.ID.String(),
		Description: res.Description,
		Priority:    priority,
		Dates: internal.Dates{
			Start: res.StartDate.Time,
			Due:   res.DueDate.Time,
		},
		IsDone: res.Done,
	}, nil
}

// Update updates the existing record with new values.
func (t *Task) Update(ctx context.Context, id string, description string, priority internal.Priority, dates internal.Dates, isDone bool) error {
	defer newOTELSpan(ctx, "Task.Update").End()

	//-

	// XXX: We will revisit the number of received arguments in future episodes.
	val, err := uuid.Parse(id)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid uuid")
	}

	if _, err := t.q.UpdateTask(ctx, db.UpdateTaskParams{
		ID:          val,
		Description: description,
		Priority:    newPriority(priority),
		StartDate:   newTimestamp(dates.Start),
		DueDate:     newTimestamp(dates.Due),
		Done:        isDone,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return internal.WrapErrorf(err, internal.ErrorCodeNotFound, "task not found")
		}

		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "update task")
	}

	return nil
}
