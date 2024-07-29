package main

import (
	"github.com/watchedsky-social/backend/pkg/database/model"
	"gorm.io/gen"
)

type CustomZoneQueries interface {
	// SELECT count(*) FROM zones WHERE ST_Intersects(border, ST_SetSRID(ST_MakeBox2D(@southEast, @northWest), 4326));
	CountVisibleZones(southEast model.Geometry, northWest model.Geometry) (int64, error)

	// SELECT * FROM zones WHERE ST_Intersects(border, ST_SetSRID(ST_MakeBox2D(@southEast, @northWest), 4326)) ORDER BY concat(name, ' ', type, ' ', state) LIMIT 10;
	ShowVisibleZones(southEast model.Geometry, northWest model.Geometry) ([]*gen.T, error)

	// SELECT id FROM zones;
	ListIDs() ([]string, error)
}

type CustomMapSearchQueries interface {
	// SELECT * FROM mapsearch WHERE to_tsvector('english', name) \@\@ to_tsquery('english', @searchText || ':*') ORDER BY ST_DistanceSphere(centroid, @mapcenter) LIMIT 10;
	PrefixSearch(searchText string, mapcenter model.Geometry) ([]*gen.T, error)
}
