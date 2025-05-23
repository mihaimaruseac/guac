//
// Copyright 2023 The GUAC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CertifyVex holds the schema definition for the CertifyVex entity.
type CertifyVex struct {
	ent.Schema
}

// Fields of the VEX.
func (CertifyVex) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(getUUIDv7).
			Unique().
			Immutable(),
		field.UUID("package_id", getUUIDv7()).Optional().Nillable(),
		field.UUID("artifact_id", getUUIDv7()).Optional().Nillable(),
		field.UUID("vulnerability_id", getUUIDv7()),
		field.Time("known_since"),
		field.String("status"),
		field.String("statement"),
		field.String("status_notes"),
		field.String("justification"),
		field.String("origin"),
		field.String("collector"),
		field.String("document_ref"),
	}
}

// Edges of the VEX.
func (CertifyVex) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("package", PackageVersion.Type).Unique().Field("package_id"),
		edge.To("artifact", Artifact.Type).Unique().Field("artifact_id"),
		edge.To("vulnerability", VulnerabilityID.Type).Unique().Required().Field("vulnerability_id"),
	}
}

// Indexes of the VEX.
func (CertifyVex) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("known_since", "justification", "status", "origin", "collector", "document_ref").
			Edges("vulnerability", "package").Unique().Annotations(entsql.IndexWhere("artifact_id IS NULL")).StorageKey("vex_artifact_id"),
		index.Fields("known_since", "justification", "status", "origin", "collector", "document_ref").
			Edges("vulnerability", "artifact").Unique().Annotations(entsql.IndexWhere("package_id IS NULL")).StorageKey("vex_package_id"),
	}
}
