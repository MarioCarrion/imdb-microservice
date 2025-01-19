package postgresql_test

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/tern/v2/migrate"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/postgresql"
)

func TestTask_Create(t *testing.T) {
	t.Parallel()

	t.Run("Create: OK", func(t *testing.T) {
		t.Parallel()

		task, err := postgresql.NewTask(newDB(t)).Create(context.Background(),
			internal.CreateParams{
				Description: "test",
				Priority:    internal.PriorityNone,
				Dates:       internal.Dates{},
			})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		if task.ID == "" {
			t.Fatalf("expected valid record, got empty value")
		}
	})

	t.Run("Create: ERR", func(t *testing.T) {
		t.Parallel()

		_, err := postgresql.NewTask(newDB(t)).Create(context.Background(),
			internal.CreateParams{
				Description: "",
				Priority:    internal.Priority(-1),
				Dates:       internal.Dates{},
			})
		if err == nil { // because of invalid priority
			t.Fatalf("expected error, got no value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeUnknown {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})
}

func TestTask_Delete(t *testing.T) {
	t.Parallel()

	t.Run("Delete: OK", func(t *testing.T) {
		t.Parallel()

		store := postgresql.NewTask(newDB(t))

		createdTask, err := store.Create(context.Background(), internal.CreateParams{
			Description: "test",
			Priority:    internal.PriorityNone,
			Dates:       internal.Dates{},
		})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		if err := store.Delete(context.Background(), createdTask.ID); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		if _, err = store.Find(context.Background(), createdTask.ID); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected no error, got %s", err)
		}
	})

	t.Run("Update: ERR uuid", func(t *testing.T) {
		t.Parallel()

		err := postgresql.NewTask(newDB(t)).Delete(context.Background(), "x")

		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeInvalidArgument {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})

	t.Run("Delete: ERR not found", func(t *testing.T) {
		t.Parallel()

		err := postgresql.NewTask(newDB(t)).Delete(context.Background(), "44633fe3-b039-4fb3-a35f-a57fe3c906c7")

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeNotFound {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})
}

func TestTask_Find(t *testing.T) {
	t.Parallel()

	t.Run("Find: OK", func(t *testing.T) {
		t.Parallel()

		store := postgresql.NewTask(newDB(t))

		originalTask, err := store.Create(context.Background(), internal.CreateParams{
			Description: "test",
			Priority:    internal.PriorityNone,
			Dates:       internal.Dates{},
		})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		actualTask, err := store.Find(context.Background(), originalTask.ID)
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		if !cmp.Equal(originalTask, actualTask) {
			t.Fatalf("expected result does not match: %s", cmp.Diff(originalTask, actualTask))
		}
	})

	t.Run("Find: ERR uuid", func(t *testing.T) {
		t.Parallel()

		_, err := postgresql.NewTask(newDB(t)).Find(context.Background(), "x")
		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeInvalidArgument {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})

	t.Run("Find: ERR not found", func(t *testing.T) {
		t.Parallel()

		_, err := postgresql.NewTask(newDB(t)).Find(context.Background(), "44633fe3-b039-4fb3-a35f-a57fe3c906c7")
		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeNotFound {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})
}

func TestTask_Update(t *testing.T) {
	t.Parallel()

	t.Run("Update: OK", func(t *testing.T) {
		t.Parallel()

		store := postgresql.NewTask(newDB(t))

		originalTask, err := store.Create(context.Background(), internal.CreateParams{
			Description: "test",
			Priority:    internal.PriorityNone,
			Dates:       internal.Dates{},
		})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		originalTask.Description = "changed"
		originalTask.Dates.Due = time.Now().UTC()
		originalTask.Priority = internal.PriorityHigh

		if err := store.Update(context.Background(),
			originalTask.ID,
			originalTask.Description,
			originalTask.Priority,
			originalTask.Dates,
			originalTask.IsDone); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		actualTask, err := store.Find(context.Background(), originalTask.ID)
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		opts := cmp.Comparer(func(x, y time.Time) bool {
			return x.Unix() == y.Unix()
		})

		if !cmp.Equal(originalTask, actualTask, opts) {
			t.Fatalf("expected result does not match: %s", cmp.Diff(originalTask, actualTask))
		}
	})

	t.Run("Update: ERR uuid", func(t *testing.T) {
		t.Parallel()

		err := postgresql.NewTask(newDB(t)).Update(context.Background(),
			"x",
			"",
			internal.PriorityNone,
			internal.Dates{},
			false)
		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeInvalidArgument {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})

	t.Run("Update: ERR invalid priority", func(t *testing.T) {
		t.Parallel()

		store := postgresql.NewTask(newDB(t))

		task, err := store.Create(context.Background(), internal.CreateParams{
			Description: "test",
			Priority:    internal.PriorityNone,
			Dates:       internal.Dates{},
		})
		if err != nil {
			t.Fatalf("expected no error, got %s", err)
		}

		err = postgresql.NewTask(newDB(t)).Update(context.Background(),
			task.ID,
			"",
			internal.Priority(-1),
			internal.Dates{},
			false)
		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeUnknown {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})

	t.Run("Update: ERR not found", func(t *testing.T) {
		t.Parallel()

		err := postgresql.NewTask(newDB(t)).Update(context.Background(),
			"44633fe3-b039-4fb3-a35f-a57fe3c906c7",
			"",
			internal.PriorityNone,
			internal.Dates{},
			false)
		if err == nil {
			t.Fatalf("expected error, got not value")
		}

		var ierr *internal.Error
		if !errors.As(err, &ierr) || ierr.Code() != internal.ErrorCodeNotFound {
			t.Fatalf("expected %T error, got %T : %v", ierr, err, err)
		}
	})
}

func newDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()

	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("username", "password"),
		Path:   "todo",
	}

	q := dsn.Query()
	q.Add("sslmode", "disable")

	dsn.RawQuery = q.Encode()

	//-

	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("Couldn't connect to docker: %s", err)
	}

	pool.MaxWait = 5 * time.Second

	pw, _ := dsn.User.Password()

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16.2-bullseye",
		Env: []string{
			"POSTGRES_USER=" + dsn.User.Username(),
			"POSTGRES_PASSWORD=" + pw,
			"POSTGRES_DB=" + dsn.Path,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		tb.Fatalf("Couldn't start resource: %s", err)
	}

	_ = resource.Expire(60)

	tb.Cleanup(func() {
		if errC := pool.Purge(resource); errC != nil {
			tb.Fatalf("Couldn't purge container: %v", errC)
		}
	})

	dsn.Host = resource.Container.NetworkSettings.IPAddress + ":5432"
	if runtime.GOOS == "darwin" { // MacOS-specific
		dsn.Host = net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))
	}

	ctx, cFunc := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	tb.Cleanup(cFunc)

	var db *pgx.Conn

	for range 20 {
		db, err = pgx.Connect(ctx, dsn.String())
		if err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if db == nil {
		tb.Fatalf("Couldn't connect to database: %s", err)
	}

	defer db.Close(ctx)

	if err = pool.Retry(func() (err error) {
		return db.Ping(ctx)
	}); err != nil {
		tb.Fatalf("Couldn't ping DB: %s", err)
	}

	//-

	migrator, err := migrate.NewMigrator(ctx, db, "public.schema_version")
	if err != nil {
		tb.Fatalf("Couldn't migrate (1): %s", err)
	}

	err = migrator.LoadMigrations(os.DirFS("../../db/migrations/"))
	if err != nil {
		tb.Fatalf("Couldn't migrate (2): %s", err)
	}

	if err = migrator.Migrate(context.Background()); err != nil {
		tb.Fatalf("Couldnt' migrate (3): %s", err)
	}

	//-

	dbpool, err := pgxpool.New(context.Background(), dsn.String())
	if err != nil {
		tb.Fatalf("Couldn't open DB Pool: %s", err)
	}

	tb.Cleanup(func() {
		dbpool.Close()
	})

	return dbpool
}
