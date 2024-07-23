//go:build loaders
// +build loaders

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/watchedsky-social/backend/pkg/cli"
	"github.com/watchedsky-social/backend/pkg/database/loaders"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var args cli.Args
	kong.Parse(&args)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s TimeZone=UTC", args.Host, args.Username, args.Password, args.DB)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.Create(path.Join(cwd, "dbloader.log"))
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetPrefix("[dbloader] ")
	log.SetFlags(log.LstdFlags)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.Default(), logger.Config{
			LogLevel: logger.Error,
		}),
	})
	if err != nil {
		log.Fatal(err)
	}

	query.SetDefault(db)
	exitCode := 0

	tx := query.Q.Begin()
	log.Println("Loading NWS Zone data into DB...")
	if err = loaders.LoadNWSZones(context.Background(), tx); err != nil {
		log.Println(err)
		tx.Rollback()
		exitCode = 1
	} else {
		tx.Commit()
	}

	tx = query.Q.Begin()
	log.Println("Loading Map Search data into DB...")
	if err = loaders.LoadMapSearchData(context.Background(), tx); err != nil {
		log.Println(err)
		tx.Rollback()
		exitCode = 2
	} else {
		tx.Commit()
	}

	os.Exit(exitCode)
}
