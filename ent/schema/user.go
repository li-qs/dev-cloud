package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			MaxLen(64).
			Unique(),

		field.String("password").
			MaxLen(255),

		field.String("nickname").
			MaxLen(64).
			Optional(),

		field.Int8("status").
			Default(1),

		field.Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

func (User) Annotation() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "user",
		},
	}
}

func (User) Index() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
	}
}
