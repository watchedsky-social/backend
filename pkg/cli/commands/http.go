package commands

import (
	"context"
	"fmt"

	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/watchedsky-social/backend"
	"github.com/watchedsky-social/backend/pkg/cli/args"
	"github.com/watchedsky-social/backend/pkg/handlers"
)

type HTTPCommand struct {
	Port args.NonrootPort `short:"p" default:"8000" help:"The HTTP listening port"`
}

func (h *HTTPCommand) Run(ctx context.Context, production bool) error {
	app := fiber.New(fiber.Config{
		Prefork:           false,
		StrictRouting:     false,
		CaseSensitive:     true,
		UnescapePath:      true,
		EnablePrintRoutes: true,
		GETOnly:           true,
		BodyLimit:         -1,
		ServerHeader:      "watchedsky",
		AppName:           "WatchedSky",
		Network:           fiber.NetworkTCP,
	})

	middlewares := []any{
		compress.New(),
	}

	if !production {
		middlewares = append(middlewares, cors.New())
	}

	middlewares = append(middlewares,
		recover.New(),
		requestid.New(),
		logger.New(),
		helmet.New(),
	)

	app.Use(middlewares...)

	api := app.Group("/api/v1")
	api.Get("/typeahead", handlers.Typeahead)
	api.Get("/zones/visible", handlers.VisibleZones)

	if production {
		app.Get("/*", filesystem.New(filesystem.Config{
			Root:       http.FS(backend.FrontendFS),
			PathPrefix: "frontend",
		}))
	}

	go func() {
		app.Listen(fmt.Sprintf(":%d", h.Port))
	}()

	<-ctx.Done()
	return app.Shutdown()
}
