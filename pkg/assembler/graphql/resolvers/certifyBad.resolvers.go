package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.56

import (
	"context"

	"github.com/guacsec/guac/pkg/assembler/graphql/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// IngestCertifyBad is the resolver for the ingestCertifyBad field.
func (r *mutationResolver) IngestCertifyBad(ctx context.Context, subject model.PackageSourceOrArtifactInput, pkgMatchType model.MatchFlags, certifyBad model.CertifyBadInputSpec) (string, error) {
	funcName := "IngestCertifyBad"
	if err := validatePackageSourceOrArtifactInput(&subject, funcName); err != nil {
		return "", gqlerror.Errorf("%v ::  %s", funcName, err)
	}
	if certifyBad.KnownSince.IsZero() {
		return "", gqlerror.Errorf("certifyBad.KnownSince is a zero time")
	}
	return r.Backend.IngestCertifyBad(ctx, subject, &pkgMatchType, certifyBad)
}

// IngestCertifyBads is the resolver for the ingestCertifyBads field.
func (r *mutationResolver) IngestCertifyBads(ctx context.Context, subjects model.PackageSourceOrArtifactInputs, pkgMatchType model.MatchFlags, certifyBads []*model.CertifyBadInputSpec) ([]string, error) {
	funcName := "IngestCertifyBads"
	valuesDefined := 0
	ingestedCertifyBadsIDS := []string{}
	if len(subjects.Packages) > 0 {
		if len(subjects.Packages) != len(certifyBads) {
			return ingestedCertifyBadsIDS, gqlerror.Errorf("%v :: uneven packages and certifyBads for ingestion", funcName)
		}
		valuesDefined = valuesDefined + 1
	}
	if len(subjects.Artifacts) > 0 {
		if len(subjects.Artifacts) != len(certifyBads) {
			return ingestedCertifyBadsIDS, gqlerror.Errorf("%v :: uneven artifacts and certifyBads for ingestion", funcName)
		}
		valuesDefined = valuesDefined + 1
	}
	if len(subjects.Sources) > 0 {
		if len(subjects.Sources) != len(certifyBads) {
			return ingestedCertifyBadsIDS, gqlerror.Errorf("%v :: uneven sources and certifyBads for ingestion", funcName)
		}
		valuesDefined = valuesDefined + 1
	}
	if valuesDefined != 1 {
		return ingestedCertifyBadsIDS, gqlerror.Errorf("%v :: must specify at most packages, artifacts or sources", funcName)
	}

	for _, certifyBad := range certifyBads {
		if certifyBad.KnownSince.IsZero() {
			return ingestedCertifyBadsIDS, gqlerror.Errorf("certifyBads contains a zero time")
		}
	}

	return r.Backend.IngestCertifyBads(ctx, subjects, &pkgMatchType, certifyBads)
}

// CertifyBad is the resolver for the CertifyBad field.
func (r *queryResolver) CertifyBad(ctx context.Context, certifyBadSpec model.CertifyBadSpec) ([]*model.CertifyBad, error) {
	if err := validatePackageSourceOrArtifactQueryFilter(certifyBadSpec.Subject); err != nil {
		return nil, gqlerror.Errorf("CertifyBad :: %s", err)
	}
	return r.Backend.CertifyBad(ctx, &certifyBadSpec)
}

// CertifyBadList is the resolver for the CertifyBadList field.
func (r *queryResolver) CertifyBadList(ctx context.Context, certifyBadSpec model.CertifyBadSpec, after *string, first *int) (*model.CertifyBadConnection, error) {
	if err := validatePackageSourceOrArtifactQueryFilter(certifyBadSpec.Subject); err != nil {
		return nil, gqlerror.Errorf("CertifyBad :: %s", err)
	}
	return r.Backend.CertifyBadList(ctx, certifyBadSpec, after, first)
}
