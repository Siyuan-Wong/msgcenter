// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 32},
		{Name: "name", Type: field.TypeString, Size: 32},
		{Name: "password", Type: field.TypeString, Nullable: true, Size: 32},
		{Name: "delete_flag", Type: field.TypeBool, Default: false},
		{Name: "gender", Type: field.TypeEnum, Enums: []string{"MAN", "WOMAN", "OTHER"}, Default: "OTHER"},
		{Name: "phone_number", Type: field.TypeString, Nullable: true, Size: 32},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		UsersTable,
	}
)

func init() {
}
