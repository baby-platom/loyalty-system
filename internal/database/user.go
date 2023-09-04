package database

import (
	"errors"

	"gorm.io/gorm"
)

func FillUserByID(user *User, id uint) (err error) {
	userFilter := User{CustomBaseModel: CustomBaseModel{ID: id}}
	res := DB.Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return res.Error
	}
	return err
}
