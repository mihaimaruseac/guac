package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.56

import (
	"context"

	"github.com/guacsec/guac/pkg/assembler/graphql/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// IngestHashEqual is the resolver for the ingestHashEqual field.
func (r *mutationResolver) IngestHashEqual(ctx context.Context, artifact model.IDorArtifactInput, otherArtifact model.IDorArtifactInput, hashEqual model.HashEqualInputSpec) (string, error) {
	return r.Backend.IngestHashEqual(ctx, artifact, otherArtifact, hashEqual)
}

// IngestHashEquals is the resolver for the ingestHashEquals field.
func (r *mutationResolver) IngestHashEquals(ctx context.Context, artifacts []*model.IDorArtifactInput, otherArtifacts []*model.IDorArtifactInput, hashEquals []*model.HashEqualInputSpec) ([]string, error) {
	funcName := "IngestHashEquals"
	ingestedHashEqualsIDS := []string{}
	if len(artifacts) != len(otherArtifacts) {
		return ingestedHashEqualsIDS, gqlerror.Errorf("%v :: uneven artifacts and other artifacts for ingestion", funcName)
	} else if len(artifacts) != len(hashEquals) {
		return ingestedHashEqualsIDS, gqlerror.Errorf("%v :: uneven artifacts and hashEquals for ingestion", funcName)
	}

	return r.Backend.IngestHashEquals(ctx, artifacts, otherArtifacts, hashEquals)
}

// HashEqual is the resolver for the HashEqual field.
func (r *queryResolver) HashEqual(ctx context.Context, hashEqualSpec model.HashEqualSpec) ([]*model.HashEqual, error) {
	if hashEqualSpec.Artifacts != nil && len(hashEqualSpec.Artifacts) > 2 {
		return nil, gqlerror.Errorf("HashEqual :: Provided spec has too many Artifacts")
	}
	return r.Backend.HashEqual(ctx, &hashEqualSpec)
}

// HashEqualList is the resolver for the HashEqualList field.
func (r *queryResolver) HashEqualList(ctx context.Context, hashEqualSpec model.HashEqualSpec, after *string, first *int) (*model.HashEqualConnection, error) {
	if hashEqualSpec.Artifacts != nil && len(hashEqualSpec.Artifacts) > 2 {
		return nil, gqlerror.Errorf("HashEqual :: Provided spec has too many Artifacts")
	}
	return r.Backend.HashEqualList(ctx, hashEqualSpec, after, first)
}
