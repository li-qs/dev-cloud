package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RefreshToken holds the schema definition for the RefreshToken entity.
type RefreshToken struct {
	ent.Schema
}

// Fields of the RefreshToken.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),

		field.String("token_hash").
			MaxLen(64).
			Unique(),

		field.Time("expires_at"),

		field.Time("created_at").
			Default(time.Now),

		field.Time("revoked_at").
			Optional().
			Nillable(),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return nil
}

func (RefreshToken) Annotation() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "refresh_token",
		},
	}
}

func (RefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("token_hash"),
	}
}
