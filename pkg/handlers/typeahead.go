package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/paulmach/orb"
	"github.com/watchedsky-social/backend/pkg/database/model"
	"github.com/watchedsky-social/backend/pkg/database/query"
)

func Typeahead(ctx *fiber.Ctx) error {
	prefix := ctx.Query("prefix")
	lon := ctx.QueryFloat("lon")
	lat := ctx.QueryFloat("lat")
	zones, err := query.Mapsearch.WithContext(ctx.UserContext()).PrefixSearch(prefix, model.NewGenericGeometry(orb.Point{lon, lat}))
	if err != nil {
		return err
	}

	return ctx.JSON(zones)
}
