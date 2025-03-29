package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserDeviceRelation 定义用户设备关联表结构
type UserDeviceRelation struct {
	ent.Schema
}

// Fields of the UserDeviceRelation.
func (UserDeviceRelation) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32),
		field.String("user_id").
			MaxLen(32),
		field.String("device_id").
			MaxLen(32),
		field.String("tag").
			MaxLen(32).
			Optional(),
	}
}

// Edges of the UserDeviceRelation.
func (UserDeviceRelation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("user_device_relations").
			Field("user_id").
			Required().
			Unique(),
		edge.From("device", Device.Type).
			Ref("user_device_relations").
			Field("device_id").
			Required().
			Unique(),
	}
}
