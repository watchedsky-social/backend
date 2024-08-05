package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/paulmach/orb"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

const maxZoneReturn = 20

var savedIDKeySet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-")

func VisibleZones(ctx *fiber.Ctx) error {
	sePoint := ctx.Query("boxse")
	nwPoint := ctx.Query("boxnw")

	seSlice := strings.Split(sePoint, ",")
	nwSlice := strings.Split(nwPoint, ",")

	seLon, err := strconv.ParseFloat(seSlice[0], 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	seLat, err := strconv.ParseFloat(seSlice[1], 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	nwLon, err := strconv.ParseFloat(nwSlice[0], 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	nwLat, err := strconv.ParseFloat(nwSlice[1], 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}

	se := model.NewGenericGeometry(orb.Point{seLon, seLat})
	nw := model.NewGenericGeometry(orb.Point{nwLon, nwLat})

	dao := query.Zone.WithContext(ctx.UserContext())
	zoneCount, err := dao.CountVisibleZones(se, nw)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}

	if zoneCount == 0 {
		return ctx.SendStatus(http.StatusNoContent)
	}

	if zoneCount > maxZoneReturn {
		return ctx.SendStatus(http.StatusUnprocessableEntity)
	}

	zones, err := query.Zone.WithContext(ctx.UserContext()).ShowVisibleZones(se, nw)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}

	return ctx.JSON(zones)
}

func GetWatchID(ctx *fiber.Ctx) error {
	zones := strings.Split(strings.ToUpper(ctx.Query("zones")), ",")
	sort.Strings(zones)

	longIDs := longID(zones...)

	passedZones := strings.Join(zones, ",")
	allZoneIDs, err := query.Zone.WithContext(ctx.UserContext()).FindCongruentZones(longIDs)
	if err != nil {
		return handleWatchIDError(ctx, err)
	}

	shortIDs := sort.StringSlice{}
	for _, id := range allZoneIDs {
		shortIDs = append(shortIDs, path.Base(id))
	}
	sort.Sort(shortIDs)

	sa := query.SavedArea
	savedArea, err := sa.WithContext(ctx.UserContext()).Where(sa.PassedZones.Eq(strings.Join(zones, ","))).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return handleWatchIDError(ctx, err)
		}
	}

	if savedArea == nil {
		merged, err := mergeGeometries(ctx.UserContext(), longIDs)
		if err != nil {
			return handleWatchIDError(ctx, err)

		}

		savedArea = &model.SavedArea{
			ID:              generateID(12),
			PassedZones:     passedZones,
			CalculatedZones: strings.Join(shortIDs, ","),
			Border:          merged,
		}

		err = sa.WithContext(ctx.UserContext()).Create(savedArea)
		if err != nil {
			return handleWatchIDError(ctx, err)

		}
	}

	return ctx.JSON(savedArea)
}

func generateID(idLen uint) string {
	b := &strings.Builder{}
	b.WriteRune(savedIDKeySet[rand.Intn(len(savedIDKeySet)-2)])
	for range idLen - 1 {
		b.WriteRune(savedIDKeySet[rand.Intn(len(savedIDKeySet))])
	}
	return b.String()
}

func longID(zones ...string) []string {
	longIDs := make([]string, len(zones))
	for i := range zones {
		t := "forecast"
		if zones[i][2] == 'C' {
			t = "county"
		}

		longIDs[i] = fmt.Sprintf("https://api.weather.gov/zones/%s/%s", t, zones[i])
	}

	return longIDs
}

func mergeGeometries(ctx context.Context, zones []string) (*model.Geometry, error) {
	allZones, err := query.Zone.WithContext(ctx).Select(field.ALL).Where(query.Zone.ID.In(zones...)).Find()
	if err != nil {
		return nil, err
	}

	mp := orb.MultiPolygon{}
	for _, z := range allZones {
		mergeRecursive(z.Border.ToOrbGeometry(), &mp)
	}

	geo := model.NewGenericGeometry(mp)
	return &geo, nil
}

func mergeRecursive(zone orb.Geometry, collector *orb.MultiPolygon) {
	switch zone.GeoJSONType() {
	case "Polygon":
		p, _ := zone.(orb.Polygon)
		*collector = append(*collector, p)
	case "MultiPolygon":
		p, _ := zone.(orb.MultiPolygon)
		*collector = append(*collector, p...)
	case "GeometryCollection":
		collection, _ := zone.(orb.Collection)
		for _, coll := range collection {
			mergeRecursive(coll, collector)
		}
	}
}

func handleWatchIDError(ctx *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	ctx.Status(http.StatusInternalServerError).JSON(map[string]string{
		"error": err.Error(),
	})

	return err
}
