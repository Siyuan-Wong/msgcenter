package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Device 定义设备表结构
type Device struct {
	ent.Schema
}

// Fields of the Device.
func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32),
		field.String("client_device_id").
			MaxLen(32).
			Unique(),
		field.Time("last_active_time").
			Default(time.Now),
		field.Bool("actived").
			Default(true),
		field.Bool("delete_flag").
			Default(false),
		field.String("curr_user_id").
			MaxLen(32).
			Optional(),
		field.Enum("device_type").
			NamedValues(
				"Mobile", "Mobile",
				"Desktop", "Desktop",
				"Other", "Other",
			).
			Default("Other"),
	}
}

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("devices").
			Field("curr_user_id").
			Unique(),
		edge.To("user_device_relations", UserDeviceRelation.Type),
	}
}
