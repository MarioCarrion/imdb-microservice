package internal

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
)

func NewMemcached(conf *envvar.Configuration) (*memcache.Client, error) {
	host, err := conf.Get("MEMCACHED_HOST")
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "conf.Get MEMCACHED_HOST")
	}

	// XXX Assuming environment variable contains only one server
	client := memcache.New(host)

	if err := client.Ping(); err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ping")
	}

	client.Timeout = 100 * time.Millisecond
	client.MaxIdleConns = 100

	return client, nil
}
