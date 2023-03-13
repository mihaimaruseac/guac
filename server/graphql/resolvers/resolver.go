package resolvers

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"layout/server/backends"
)

type Resolver struct {
	Backend backends.Backend
}
