package userService

import "context"

type UserService interface {
	InsertDemo(ctx context.Context) error
}
