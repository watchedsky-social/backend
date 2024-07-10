package model

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
)

type Point orb.Point

const SRID = 4326

type Geometry struct {
	g orb.Geometry
}

func NewGenericGeometry(g orb.Geometry) Geometry {
	return Geometry{g: g}
}

func (g Geometry) ToOrbGeometry() orb.Geometry {
	return g.g
}

func (g *Geometry) Scan(src any) error {
	// PostGIS stores geometry columns as ewkb bytes.
	// if src is a []byte use it. If it's a string, it's hex-encoded.
	var (
		ewkbData []byte
		err      error
	)
	switch data := src.(type) {
	case []byte:
		ewkbData = data
	case string:
		ewkbData, err = hex.DecodeString(data)
		if err != nil {
			return fmt.Errorf("could not decode gis data: %w", err)
		}
	default:
		return fmt.Errorf("unexpected data type %T for src", data)
	}

	geo, _, err := ewkb.Unmarshal(ewkbData)
	if err != nil {
		return err
	}

	*g = Geometry{g: geo}
	return nil
}

func (g Geometry) Value() (driver.Value, error) {
	return ewkb.MarshalToHex(g.g, SRID)
}

func (g *Geometry) UnmarshalJSON(data []byte) error {
	var geo orb.Geometry
	if err := json.Unmarshal(data, &geo); err != nil {
		return err
	}

	*g = Geometry{g: geo}
	return nil
}

func (g Geometry) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.g)
}
