package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

func GetUserByID(ctx context.Context, id uint) (user User, err error) {
	userFilter := User{CustomBaseModel: CustomBaseModel{ID: id}}
	res := DB.Conn(ctx).Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return user, res.Error
	}
	return
}

func createUser(ctx context.Context, login string, hash string) (User, error) {
	user := User{Login: login, PasswordHash: hash}
	return user, DB.Conn(ctx).Create(&user).Error
}

func CreateUserWithBalance(ctx context.Context, login string, hash string) (user User, err error) {
	user, err = createUser(ctx, login, hash)
	if err != nil {
		return
	}

	_, err = createBalance(ctx, user.ID)
	return
}

func GetUserByLogin(ctx context.Context, login string) (user User, err error) {
	filter := User{Login: login}
	return user, DB.Conn(ctx).Where(&filter).First(&user).Error
}
