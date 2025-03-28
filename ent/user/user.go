// Code generated by ent, DO NOT EDIT.

package user

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldPassword holds the string denoting the password field in the database.
	FieldPassword = "password"
	// FieldDeleteFlag holds the string denoting the delete_flag field in the database.
	FieldDeleteFlag = "delete_flag"
	// FieldGender holds the string denoting the gender field in the database.
	FieldGender = "gender"
	// FieldPhoneNumber holds the string denoting the phone_number field in the database.
	FieldPhoneNumber = "phone_number"
	// Table holds the table name of the user in the database.
	Table = "users"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldPassword,
	FieldDeleteFlag,
	FieldGender,
	FieldPhoneNumber,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// PasswordValidator is a validator for the "password" field. It is called by the builders before save.
	PasswordValidator func(string) error
	// DefaultDeleteFlag holds the default value on creation for the "delete_flag" field.
	DefaultDeleteFlag bool
	// PhoneNumberValidator is a validator for the "phone_number" field. It is called by the builders before save.
	PhoneNumberValidator func(string) error
	// IDValidator is a validator for the "id" field. It is called by the builders before save.
	IDValidator func(string) error
)

// Gender defines the type for the "gender" enum field.
type Gender string

// GenderOTHER is the default value of the Gender enum.
const DefaultGender = GenderOTHER

// Gender values.
const (
	GenderMAN   Gender = "MAN"
	GenderWOMAN Gender = "WOMAN"
	GenderOTHER Gender = "OTHER"
)

func (ge Gender) String() string {
	return string(ge)
}

// GenderValidator is a validator for the "gender" field enum values. It is called by the builders before save.
func GenderValidator(ge Gender) error {
	switch ge {
	case GenderMAN, GenderWOMAN, GenderOTHER:
		return nil
	default:
		return fmt.Errorf("user: invalid enum value for gender field: %q", ge)
	}
}

// OrderOption defines the ordering options for the User queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByPassword orders the results by the password field.
func ByPassword(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPassword, opts...).ToFunc()
}

// ByDeleteFlag orders the results by the delete_flag field.
func ByDeleteFlag(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeleteFlag, opts...).ToFunc()
}

// ByGender orders the results by the gender field.
func ByGender(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldGender, opts...).ToFunc()
}

// ByPhoneNumber orders the results by the phone_number field.
func ByPhoneNumber(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPhoneNumber, opts...).ToFunc()
}
