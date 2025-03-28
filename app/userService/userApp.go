package userService

import (
	"context"
	"msgcenter/app"
	"msgcenter/ent/user"
	"msgcenter/utils/objectid"
	"sync"
)

var (
	instance *UserApp
	once     sync.Once
)

type UserApp struct {
	s *app.ServiceApp
}

func USER() *UserApp {
	once.Do(func() {
		instance = &UserApp{
			s: app.SERVICE(),
		}
	})
	return instance
}

func (userApp *UserApp) InsertDemo(ctx context.Context) error {
	// 插入数据
	// 示例：
	_, err := userApp.s.DbClient.User.Create().
		SetName("John Doe").
		SetID(objectid.New()).
		SetGender(user.GenderOTHER).
		SetPhoneNumber("12345678900").
		Save(ctx)
	if err != nil {
		return err
	}
	return nil
}
