package db

import (
	"fmt"
	"os"
)

var models []any
var migrations []string

func RegisterModel(model any) {
	models = append(models, model)
}

func RegisterMigration(migration string) {
	migrations = append(migrations, migration)
}

func autoMigrate() {
	for _, migration := range migrations {
		Conn.Exec(migration)
	}

	if os.Getenv("APP_ENV") != "dev" {
		if os.Getenv("FORCE_MIGRATE") == "yes" && os.Getenv("APP_ENV") != "prod" {
			err := Conn.Migrator().DropTable(models...)
			if err != nil {
				panic(err)
			}
			fmt.Println("Table Dropped Success....")
		}
		err := Conn.AutoMigrate(models...)
		if err != nil {
			panic(err)
		}
		fmt.Println("Table Migrated Success....")
	}
}
