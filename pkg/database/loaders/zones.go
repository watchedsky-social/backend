//go:build loaders
// +build loaders

package loaders

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"

	"github.com/paulmach/orb/geojson"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"github.com/watchedsky-social/backend/pkg/utils"
	"gorm.io/gorm/clause"
)

//go:embed zones.json.gz
var zonesJSONGZ []byte

func LoadNWSZones(ctx context.Context, tx *query.QueryTx) error {
	reader, err := gzip.NewReader(bytes.NewBuffer(zonesJSONGZ))
	if err != nil {
		return err
	}
	defer reader.Close()

	var fc geojson.FeatureCollection
	if err = json.NewDecoder(reader).Decode(&fc); err != nil {
		return err
	}

	zones := utils.Map(fc.Features, func(feat *geojson.Feature) *model.Zone {
		return &model.Zone{
			ID:     feat.Properties.MustString("@id"),
			Type:   feat.Properties.MustString("type"),
			Name:   feat.Properties.MustString("name"),
			State:  utils.Ref(feat.Properties.MustString("state", "")),
			Border: model.NewGenericGeometry(feat.Geometry),
		}
	})

	dao := tx.WithContext(ctx).Zone.Clauses(clause.OnConflict{DoNothing: true})
	for start := 0; start < len(zones); start += batchSize {
		batch := zones[start:utils.Min(start+batchSize, len(zones))]
		if err := dao.CreateInBatches(batch, len(batch)); err != nil {
			return err
		}
	}

	return nil
}

// func LoadNWSZones(ctx context.Context, tx *query.QueryTx) error {
// 	zone := tx.WithContext(ctx).Zone

// 	existingIDs, err := zone.ListIDs()
// 	if err != nil {
// 		log.Println(err)
// 		existingIDs = []string{}
// 	}

// 	ids = utils.Filter(ids, func(id string) bool {
// 		return !slices.Contains(existingIDs, id)
// 	})

// 	wg := &sync.WaitGroup{}
// 	wg.Add(len(ids))

// 	workQueue := make(chan string, numWorkers)
// 	bar := progressbar.Default(int64(len(ids)), "Loading zones into db...")
// 	for range numWorkers {
// 		go worker(wg, workQueue, zone, bar)
// 	}

// 	for _, id := range ids {
// 		workQueue <- id
// 	}

// 	close(workQueue)
// 	wg.Wait()

// 	return nil
// }

// func worker(wg *sync.WaitGroup, workQueue chan string, dao query.IZoneDo, bar *progressbar.ProgressBar) {
// 	zones := []*model.Zone{}
// 	for id := range workQueue {
// 		zone, err := loadZone(wg, id, bar)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		zones = append(zones, zone)
// 	}

// 	if err := dao.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(zones, len(zones)); err != nil {
// 		log.Println(err)
// 	}
// }

// func loadZone(wg *sync.WaitGroup, id string, bar *progressbar.ProgressBar) (*model.Zone, error) {
// 	defer func() {
// 		wg.Done()
// 		bar.Add(1)
// 	}()

// , nil
// }
