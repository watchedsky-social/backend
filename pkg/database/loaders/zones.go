package loaders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/paulmach/orb/geojson"
	"github.com/schollz/progressbar/v3"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"github.com/watchedsky-social/backend/pkg/utils"
	"gorm.io/gorm/clause"
)

const urlTemplate = "https://api.weather.gov/zones/%s"
const numWorkers = 50

func LoadNWSZones(ctx context.Context, tx *query.QueryTx) error {
	zone := tx.WithContext(ctx).Zone

	existingIDs, err := zone.ListIDs()
	if err != nil {
		log.Println(err)
		existingIDs = []string{}
	}

	modes := []string{"county", "forecast"}

	ids := make([]string, 0, 10_000)
	for _, m := range modes {
		url := fmt.Sprintf(urlTemplate, m)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal(fmt.Errorf("expected 200 OK, got %s", resp.Status))
		}

		var fc geojson.FeatureCollection
		if err = json.NewDecoder(resp.Body).Decode(&fc); err != nil {
			return err
		}

		ids = append(ids, utils.Map(fc.Features, func(f *geojson.Feature) string {
			return f.ID.(string)
		})...)
	}

	ids = utils.Filter(ids, func(id string) bool {
		return !slices.Contains(existingIDs, id)
	})

	wg := &sync.WaitGroup{}
	wg.Add(len(ids))

	workQueue := make(chan string, numWorkers)
	bar := progressbar.Default(int64(len(ids)), "Loading zones into db...")
	for range numWorkers {
		go worker(wg, workQueue, zone, bar)
	}

	for _, id := range ids {
		workQueue <- id
	}

	close(workQueue)
	wg.Wait()

	return nil
}

func worker(wg *sync.WaitGroup, workQueue chan string, dao query.IZoneDo, bar *progressbar.ProgressBar) {
	zones := []*model.Zone{}
	for id := range workQueue {
		zone, err := loadZone(wg, id, bar)
		if err != nil {
			log.Println(err)
			continue
		}

		zones = append(zones, zone)
	}

	if err := dao.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(zones, len(zones)); err != nil {
		log.Println(err)
	}
}

func loadZone(wg *sync.WaitGroup, id string, bar *progressbar.ProgressBar) (*model.Zone, error) {
	defer func() {
		wg.Done()
		bar.Add(1)
	}()

	resp, err := http.Get(id)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 OK, got %s", resp.Status)
	}

	var feat geojson.Feature
	if err = json.NewDecoder(resp.Body).Decode(&feat); err != nil {
		return nil, err
	}

	return &model.Zone{
		ID:     feat.Properties.MustString("@id"),
		Type:   feat.Properties.MustString("type"),
		Name:   feat.Properties.MustString("name"),
		State:  utils.Ref(feat.Properties.MustString("state", "")),
		Border: model.NewGenericGeometry(feat.Geometry),
		Center: model.NewGenericGeometry(feat.Geometry.Bound().Center()),
	}, nil
}
