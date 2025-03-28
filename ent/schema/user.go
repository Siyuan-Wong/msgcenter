package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Gender 定义性别枚举
type Gender string

const (
	GenderMan   Gender = "MAN"
	GenderWoman Gender = "WOMAN"
	GenderOther Gender = "OTHER"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// User 定义用户实体
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(32).
			NotEmpty().
			Unique(),

		field.String("name").
			MaxLen(32).
			NotEmpty(),

		field.String("password").
			MaxLen(32).
			Optional(),

		field.Bool("delete_flag").
			Default(false),

		field.Enum("gender").
			Values(
				string(GenderMan),
				string(GenderWoman),
				string(GenderOther),
			).
			Default(string(GenderOther)),

		field.String("phone_number").
			MaxLen(32).
			Optional(),
	}
}

// Edges 定义关系（当前无关联关系）
func (User) Edges() []ent.Edge {
	return nil
}
