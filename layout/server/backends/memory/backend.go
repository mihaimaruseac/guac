package memory

import (
	"layout/server/backends"
)

type Credentials struct{}

type client struct{}

func Backend(args backends.BackendArgs) (backends.Backend, error) {
	c := &client{}
	return c, nil
}
