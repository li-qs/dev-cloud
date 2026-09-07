package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),

		field.Int("resource_id"),

		field.Enum("type").
			Values(
				"CREATE_RESOURCE",
				"START_RESOURCE",
				"STOP_RESOURCE",
				"RESTART_RESOURCE",
				"REMOVE_RESOURCE",
			),

		field.Enum("status").
			Values(
				"PENDING",
				"RUNNING",
				"SUCCESS",
				"FAILED",
			).
			Default("PENDING"),

		field.Int("attempts").
			Default(0),

		field.Int("max_attempts").
			Default(3),

		field.Time("next_run_at").
			Optional(),

		field.JSON("payload", map[string]any{}).
			Optional(),

		field.Text("error_message").
			Optional(),

		field.Time("started_at").
			Optional().
			Nillable(),

		field.Time("finished_at").
			Optional().
			Nillable(),

		field.Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return nil
}

func (Task) Annotation() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "task",
		},
	}
}

func (Task) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "resource_id"),
	}
}
