//go:build loaders
// +build loaders

package loaders

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/schollz/progressbar/v3"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"github.com/watchedsky-social/backend/pkg/utils"
	"gorm.io/gorm/clause"
)

//go:embed US.txt
var usTSV []byte

const batchSize = 10_000

func LoadMapSearchData(ctx context.Context, tx *query.QueryTx) error {
	reader := csv.NewReader(bytes.NewBuffer(usTSV))
	reader.Comma = '\t'
	reader.FieldsPerRecord = 12

	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	data := make([]*model.Mapsearch, 0, len(rows))
	bar := progressbar.Default(int64(len(rows)), "Loading data into mapsearch table...")
	for _, row := range rows {
		data = append(data, rowToDatum(row))
		bar.Add(1)
	}

	dao := tx.WithContext(ctx).Mapsearch.Clauses(clause.OnConflict{DoNothing: true})
	for start := 0; start < len(data); start += batchSize {
		batch := data[start:utils.Min(start+batchSize, len(data))]
		if err := dao.CreateInBatches(batch, len(batch)); err != nil {
			return err
		}
	}

	return nil
}

func rowToDatum(row []string) *model.Mapsearch {
	var (
		lat float64
		lon float64
	)

	datum := &model.Mapsearch{
		ID:        row[1],
		Name:      row[2],
		State:     row[3],
		StateName: row[4],
		County:    utils.Ref(row[5]),
	}

	lat, _ = strconv.ParseFloat(row[9], 64)
	lon, _ = strconv.ParseFloat(row[10], 64)
	datum.Centroid = model.NewGenericGeometry(orb.Point{lon, lat})

	return datum
}
