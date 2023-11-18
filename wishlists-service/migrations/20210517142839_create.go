package migrations

import (
	// "fmt"

	// "gofr.dev/gofr/pkg/datastore"
	// "gofr.dev/gofr/pkg/log"
)

// type K20210517154650 struct {
// }

// func (k K20210517154650) Up(d *datastore.DataStore, logger log.Logger) error {
// 	fmt.Println("Running migration up: 20210517153524_create.go")

// 	_,err:=d.DB().Exec("CREATE TABLE IF NOT EXISTS `users` "+
// 	    "(`id` int NOT NULL AUTO_INCREMENT, "+
// 		"`first_name` varchar(50) NOT NULL,"+
// 		"`last_name` varchar(50) NOT NULL,"+
// 		"`email_id` varchar(150) NOT NULL, "+
// 		"PRIMARY KEY (`id`))")

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (k K20210517154650) Down(d *datastore.DataStore, logger log.Logger) error {
// 	fmt.Println("Running migration down: 20210517153524_create.go")

// 	_,err:=d.DB().Exec("Drop table If EXISTS `users`")

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }