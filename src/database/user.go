package database

import (
	"github.com/Aoi-hosizora/ahlib-web/xgorm"
	"github.com/Aoi-hosizora/ahlib-web/xstatus"
	"github.com/Aoi-hosizora/scut-academic-notifier/src/model"
)

func GetUsers() []*model.User {
	users := make([]*model.User, 0)
	DB.Model(&model.User{}).Find(&users)
	return users
}

func GetUser(chatId int64) *model.User {
	user := &model.User{}
	rdb := DB.Model(&model.User{}).Where(&model.User{ChatID: chatId}).First(user)
	if rdb.RecordNotFound() {
		return nil
	}
	return user
}

func AddUser(user *model.User) xstatus.DbStatus {
	return xgorm.WithDB(DB).Insert(&model.User{}, user)
}

func DeleteUser(chatId int64) xstatus.DbStatus {
	return xgorm.WithDB(DB).Delete(&model.User{}, nil, &model.User{ChatID: chatId})
}
