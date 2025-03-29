package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User 定义用户表结构
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32),
		field.String("name").
			MaxLen(32),
		field.String("password").
			MaxLen(32).
			Optional(),
		field.Bool("delete_flag"),
		field.Enum("gender").
			NamedValues(
				"Male", "M",
				"Female", "F",
				"Other", "O",
			).
			Default("O"),
		field.String("phone_number").
			MaxLen(32).
			Optional(),
		field.Int("age").
			Default(0),
		field.String("email").
			MaxLen(64).
			Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("devices", Device.Type),
		edge.To("user_device_relations", UserDeviceRelation.Type),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("email"),
		index.Fields("phone_number"),
	}
}
