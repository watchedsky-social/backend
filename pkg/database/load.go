package database

import (
	"fmt"

	"github.com/watchedsky-social/backend/pkg/cli"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var loaded = false

func Load(dbArgs cli.DBArgs) error {
	if loaded {
		return nil
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s TimeZone=UTC",
		dbArgs.Host, dbArgs.Username, dbArgs.Password, dbArgs.DB)))
	if err != nil {
		return err
	}

	query.SetDefault(db)
	loaded = true
	return nil
}
