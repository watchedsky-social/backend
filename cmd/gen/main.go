package main

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/watchedsky-social/backend/pkg/cli"
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type CustomZoneQueries interface {
	// SELECT id FROM @@table;
	ListIDs() ([]string, error)
}

type CustomMapSearchQueries interface {
	// SELECT * FROM @@table WHERE fts_index_col \@\@ to_tsquery(@searchStr) ORDER BY ts_rank_cd(fts_index_col, to_tsquery(@searchStr), 32) DESC;
	FindBySubstring(searchStr string) ([]*gen.T, error)
}

func main() {
	var args cli.Args
	kong.Parse(&args)

	db, _ := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s TimeZone=UTC", args.Host, args.Username, args.Password, args.DB)))

	dataTypeMap := map[string]func(columnType gorm.ColumnType) (dataType string){
		"geometry": func(columnType gorm.ColumnType) (dataType string) {
			ct, _ := columnType.ColumnType()
			if strings.Contains(strings.ToLower(ct), "geometry(") {
				return "Geometry"
			}

			return "string"
		},
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:           "../../pkg/database/query",
		OutFile:           "gen_query.go",
		ModelPkgPath:      "../../pkg/database/model",
		WithUnitTest:      true,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)
	g.WithDataTypeMap(dataTypeMap)
	g.WithImportPkgPath("github.com/paulmach/orb")

	g.ApplyInterface(
		func(CustomZoneQueries) {},
		g.GenerateModel("zones"),
	)
	g.ApplyInterface(
		func(CustomMapSearchQueries) {},
		g.GenerateModel("mapsearch", gen.FieldIgnore("fts_index_col")),
	)
	g.Execute()
}
