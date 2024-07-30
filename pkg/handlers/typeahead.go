package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/watchedsky-social/backend/pkg/database/query"
)

func Typeahead(ctx *fiber.Ctx) error {
	prefix := ctx.Query("prefix")
	searchResults, err := query.Mapsearch.WithContext(ctx.UserContext()).PrefixSearch(prefix)
	if err != nil {
		return err
	}

	return ctx.JSON(searchResults)
}
