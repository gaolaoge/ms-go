package service

import (
	"fmt"
	"net/url"

	"github.com/gaolaoge/ms-go/orm"
)

type User struct {
	Id       int64
	Username string
	Password string
	Age      int
}

func SaveUser() error {
	username := "root"
	password := ""
	path := ""
	port := 3306
	database := ""
	location := url.QueryEscape("Asia/Shanghai")
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=%s&parseTime=true", username, password, path, port, database, location)

	msDb, err := orm.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}

	msDb.Prefix = "msgo_"

	user := &User{
		Id:       1000,
		Username: "gaoge",
		Password: "ps",
		Age:      18,
	}

	id, _, err := msDb.New().Table("msgo_user").Insert(user)
	if err != nil {
		return err
	}

	fmt.Println(id)

	msDb.Close()

	return nil

}
