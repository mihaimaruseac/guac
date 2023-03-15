package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Khan/genqlient/graphql"

	model "layout/client/generated"
)

func ingestData(port int) {
	// ensure server is up
	time.Sleep(1 * time.Second)

	ctx := context.Background()

	// Create a http client to send the mutation through
	url := fmt.Sprintf("http://localhost:%d/query", port)
	httpClient := http.Client{}
	gqlclient := graphql.NewClient(url, &httpClient)

	start := time.Now()
	log.Printf("Ingesting test data into backend server")
	ingestPackages(ctx, gqlclient)
	ingestSources(ctx, gqlclient)
	ingestHasSourceAt(ctx, gqlclient)
	time := time.Now().Sub(start)
	log.Printf("Ingesting test data into backend server took %v", time)
}

func ingestPackages(ctx context.Context, client graphql.Client) {
	v11 := "2.11.1"
	v12 := "2.12.0"
	subpath1 := "saved_model_cli.py"
	subpath2 := "__init__.py"
	opensslNamespace := "openssl.org"
	opensslVersion := "3.0.3"

	inputs := []model.PackageInput{{
		Type: "pypi",
		Name: "tensorflow",
	}, {
		Type:    "pypi",
		Name:    "tensorflow",
		Version: &v11,
	}, {
		Type:    "pypi",
		Name:    "tensorflow",
		Version: &v12,
	}, {
		Type:    "pypi",
		Name:    "tensorflow",
		Version: &v12,
		Subpath: &subpath1,
	}, {
		Type:    "pypi",
		Name:    "tensorflow",
		Version: &v12,
		Subpath: &subpath2,
	}, {
		Type:      "conan",
		Namespace: &opensslNamespace,
		Name:      "openssl",
		Version:   &opensslVersion,
	}}

	for _, input := range inputs {
		_, err := model.IngestPackage(ctx, client, input)
		if err != nil {
			log.Printf("Error in ingesting: %v\n", err)
		}
	}
}

func ingestSources(ctx context.Context, client graphql.Client) {
	v12 := "v2.12.0"
	commit := "abcdef"

	inputs := []model.SourceInput{{
		Type:      "git",
		Namespace: "github.com/tensorflow",
		Name:      "tensorflow",
	}, {
		Type:      "git",
		Namespace: "github.com/tensorflow",
		Name:      "build",
	}, {
		Type:      "git",
		Namespace: "github.com/tensorflow",
		Name:      "tensorflow",
		Tag:       &v12,
	}, {
		Type:      "git",
		Namespace: "github.com/tensorflow",
		Name:      "tensorflow",
		Commit:    &commit,
	}}

	for _, input := range inputs {
		_, err := model.IngestSource(ctx, client, input)
		if err != nil {
			log.Printf("Error in ingesting: %v\n", err)
		}
	}
}

func ingestHasSourceAt(ctx context.Context, client graphql.Client) {
	version := "2.12.0"
	tag := "v2.12.0"

	inputs := []struct {
		pkg   model.PackageInput
		src   model.SourceInput
		input model.HasSourceAtInput
	}{{
		pkg: model.PackageInput{
			Type:    "pypi",
			Name:    "tensorflow",
			Version: &version,
		},
		src: model.SourceInput{
			Type:      "git",
			Namespace: "github.com/tensorflow",
			Name:      "tensorflow",
			Tag:       &tag,
		},
		input: model.HasSourceAtInput{
			Justification: "TF 2.12.0 release",
		},
	}, {
		pkg: model.PackageInput{
			Type: "pypi",
			Name: "tensorflow",
		},
		src: model.SourceInput{
			Type:      "git",
			Namespace: "github.com/tensorflow",
			Name:      "tensorflow",
		},
		input: model.HasSourceAtInput{
			Justification: "General mapping between wheel and repo",
		},
	}}

	for _, input := range inputs {
		_, err := model.IngestSourceAt(ctx, client, input.pkg, input.src, input.input)
		if err != nil {
			log.Printf("Error in ingesting: %v\n", err)
		}
	}
}
