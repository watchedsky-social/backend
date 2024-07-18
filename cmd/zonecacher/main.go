package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/paulmach/orb/geojson"
	"github.com/schollz/progressbar/v3"
	"github.com/watchedsky-social/backend/pkg/utils"
)

const urlTemplate = "https://api.weather.gov/zones/%s"

func main() {
	modes := []string{"county", "forecast"}

	ids := make([]string, 0, 10_000)
	for _, m := range modes {
		url := fmt.Sprintf(urlTemplate, m)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatal(fmt.Errorf("expected 200 OK, got %s", resp.Status))
		}

		var fc geojson.FeatureCollection
		if err = json.NewDecoder(resp.Body).Decode(&fc); err != nil {
			log.Fatal(err)
		}

		ids = append(ids, utils.Map(fc.Features, func(f *geojson.Feature) string {
			return f.ID.(string)
		})...)
	}

	bar := progressbar.Default(int64(len(ids)), "Caching all zones...")

	fullFC := geojson.NewFeatureCollection()
	for _, id := range ids {
		bar.Add(1)
		resp, err := http.Get(id)
		if err != nil {
			log.Println(err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Println(fmt.Errorf("expected 200 OK, got %s", resp.Status))
			resp.Body.Close()
			continue
		}

		feat := new(geojson.Feature)
		if err = json.NewDecoder(resp.Body).Decode(feat); err != nil {
			log.Println(err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		fullFC.Append(feat)
	}

	cacheFile, err := os.CreateTemp(os.TempDir(), "zones.json.gz")
	if err != nil {
		log.Fatal(err)
	}

	writer := gzip.NewWriter(cacheFile)
	defer func() {
		writer.Close()
		cacheFile.Close()
	}()

	if err = json.NewEncoder(writer).Encode(fullFC); err != nil {
		log.Fatal(err)
	}

	fmt.Println(cacheFile.Name())
}
