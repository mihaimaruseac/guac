package backends

import (
	"context"

	"layout/server/graphql/model"
)

type Backend interface {
	// Selections
	Packages(ctx context.Context, filter model.PackageFilter) ([]*model.Package, error)
	Sources(ctx context.Context, filter model.SourceFilter) ([]*model.Source, error)

	// Ingestion
	IngestPackage(ctx context.Context, input model.PackageInput) (*model.Package, error)
	IngestSource(ctx context.Context, input model.SourceInput) (*model.Source, error)

	// Evidence tree selection
	SourceMap(ctx context.Context, filter model.HasSourceAtFilter) ([]*model.HasSourceAt, error)

	// Evidence tree ingestion
	IngestSourceAt(ctx context.Context, pkg model.PackageInput, source model.SourceInput, input model.HasSourceAtInput) (*model.HasSourceAt, error)
}

type BackendArgs interface{}
