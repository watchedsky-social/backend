package main

import (
	"github.com/watchedsky-social/backend/pkg/database/model"
	"gorm.io/gen"
)

type CustomZoneQueries interface {
	// SELECT count(*) FROM zones WHERE ST_Intersects(border, ST_SetSRID(ST_MakeBox2D(@southEast, @northWest), 4326));
	CountVisibleZones(southEast model.Geometry, northWest model.Geometry) (int64, error)

	// SELECT * FROM zones WHERE ST_Intersects(border, ST_SetSRID(ST_MakeBox2D(@southEast, @northWest), 4326)) ORDER BY concat(name, ' ', type, ' ', state) LIMIT 20;
	ShowVisibleZones(southEast model.Geometry, northWest model.Geometry) ([]*gen.T, error)

	// SELECT id FROM zones;
	ListIDs() ([]string, error)
}

type CustomMapSearchQueries interface {
	/* WITH searchResults AS (
	       SELECT * FROM mapsearch
	           WHERE display_name ILIKE @searchText || '%' OR id LIKE @searchText || '%'
	       )
	   SELECT DISTINCT ON (display_name) id, name, state, county, centroid
	       FROM searchResults ORDER by display_name;  */
	PrefixSearch(searchText string) ([]*gen.T, error)
}
