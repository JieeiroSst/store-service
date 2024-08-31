package main

import (
	"fmt"

	"github.com/JIeeiroSst/address-country-service/config"
	"github.com/JIeeiroSst/address-country-service/pkg/consul"
	"github.com/JIeeiroSst/address-country-service/pkg/mysql"
	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.New()

	dirEnv, err := config.ReadFileEnv(".env")
	if err != nil {

	}

	consul := consul.NewConfigConsul(dirEnv.HostConsul, dirEnv.KeyConsul, dirEnv.ServiceConsul)
	conf, err := consul.ConnectConfigConsul()
	if err != nil {

	}

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.Mysql.MysqlUser,
		conf.Mysql.MysqlPassword,
		conf.Mysql.MysqlHost,
		conf.Mysql.MysqlPort,
		conf.Mysql.MysqlDbname,
	)
	_ = mysql.NewMysqlConn(dns)

	app.Run()
}
