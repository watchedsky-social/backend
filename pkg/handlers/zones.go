package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/paulmach/orb"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
)

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

	zones, err := query.Zone.WithContext(ctx.UserContext()).ShowVisibleZones(se, nw)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}

	return ctx.JSON(zones)
}
