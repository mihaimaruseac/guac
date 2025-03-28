package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"

	"github.com/guacsec/guac/pkg/assembler/graphql/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// IngestOccurrence is the resolver for the ingestOccurrence field.
func (r *mutationResolver) IngestOccurrence(ctx context.Context, subject model.PackageOrSourceInput, artifact model.IDorArtifactInput, occurrence model.IsOccurrenceInputSpec) (string, error) {
	funcName := "IngestOccurrence"
	if err := validatePackageOrSourceInput(&subject, funcName); err != nil {
		return "", gqlerror.Errorf("%v :: %s", funcName, err)
	}
	return r.Backend.IngestOccurrence(ctx, subject, artifact, occurrence)
}

// IngestOccurrences is the resolver for the ingestOccurrences field.
func (r *mutationResolver) IngestOccurrences(ctx context.Context, subjects model.PackageOrSourceInputs, artifacts []*model.IDorArtifactInput, occurrences []*model.IsOccurrenceInputSpec) ([]string, error) {
	funcName := "IngestOccurrences"
	ingestedOccurrencesIDs := []string{}
	valuesDefined := 0
	if len(subjects.Packages) > 0 {
		if len(subjects.Packages) != len(artifacts) {
			return ingestedOccurrencesIDs, gqlerror.Errorf("%v :: uneven packages and artifacts for ingestion", funcName)
		}
		if len(subjects.Packages) != len(occurrences) {
			return ingestedOccurrencesIDs, gqlerror.Errorf("%v :: uneven packages and occurrence for ingestion", funcName)
		}
		valuesDefined = valuesDefined + 1
	}
	if len(subjects.Sources) > 0 {
		if len(subjects.Sources) != len(artifacts) {
			return ingestedOccurrencesIDs, gqlerror.Errorf("%v :: uneven Sources and artifacts for ingestion", funcName)
		}
		if len(subjects.Sources) != len(occurrences) {
			return ingestedOccurrencesIDs, gqlerror.Errorf("%v :: uneven Sources and occurrence for ingestion", funcName)
		}
		valuesDefined = valuesDefined + 1
	}
	if valuesDefined != 1 {
		return ingestedOccurrencesIDs, gqlerror.Errorf("%v :: must specify at most packages or sources", funcName)
	}

	return r.Backend.IngestOccurrences(ctx, subjects, artifacts, occurrences)
}

// IsOccurrence is the resolver for the IsOccurrence field.
func (r *queryResolver) IsOccurrence(ctx context.Context, isOccurrenceSpec model.IsOccurrenceSpec) ([]*model.IsOccurrence, error) {
	if err := validatePackageOrSourceQueryFilter(isOccurrenceSpec.Subject); err != nil {
		return nil, gqlerror.Errorf("IsOccurrence :: %s", err)
	}
	return r.Backend.IsOccurrence(ctx, &isOccurrenceSpec)
}

// IsOccurrenceList is the resolver for the IsOccurrenceList field.
func (r *queryResolver) IsOccurrenceList(ctx context.Context, isOccurrenceSpec model.IsOccurrenceSpec, after *string, first *int) (*model.IsOccurrenceConnection, error) {
	if err := validatePackageOrSourceQueryFilter(isOccurrenceSpec.Subject); err != nil {
		return nil, gqlerror.Errorf("IsOccurrence :: %s", err)
	}
	return r.Backend.IsOccurrenceList(ctx, isOccurrenceSpec, after, first)
}
