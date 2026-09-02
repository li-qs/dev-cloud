package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Token holds the schema definition for the Token entity.
type Token struct {
	ent.Schema
}

func (Token) Annotation() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "access_token",
		},
	}
}

// Fields of the Token.
func (Token) Fields() []ent.Field {
	return []ent.Field{
		field.Int("uid").Positive(),
		field.String("token").NotEmpty().Unique(),
		field.String("user_agent").NotEmpty(),
		field.Time("expires_at").Immutable(),
		field.Time("cerated_at").Default(time.Now),
	}
}

// Edges of the Token.
func (Token) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("access_tokens").
			Field("uid").
			Unique().
			Required(),
	}
}
