package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Resource holds the schema definition for the Resource entity.
type Resource struct {
	ent.Schema
}

// Fields of the Resource.
func (Resource) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),

		field.String("name").
			MaxLen(128),

		field.Enum("provider").
			Values("docker").
			Default("docker"),

		field.String("image").
			MaxLen(512),

		field.String("runtime_id").
			MaxLen(128).
			Optional(),

		field.Enum("status").
			Values(
				"CREATING",
				"RUNNING",
				"FAILED",
				"STOPPING",
				"STOPPED",
				"STARTING",
				"RESTARTING",
				"REMOVING",
				"REMOVED",
			).
			Default("CREATING"),

		field.JSON("config", map[string]any{}).
			Optional(),

		field.Text("credential").
			Optional(),

		field.Text("error_message").
			Optional(),

		field.Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Resource.
func (Resource) Edges() []ent.Edge {
	return nil
}

func (Resource) Annotation() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "resource",
		},
	}
}

func (Resource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}
