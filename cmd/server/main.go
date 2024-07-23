package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/watchedsky-social/backend/pkg/cli"
	"github.com/watchedsky-social/backend/pkg/database"
)

var Version string = "0.0.0-local"

func main() {
	var args cli.ServerArgs

	kongCtx := kong.Parse(&args, kong.Vars{
		"version": Version,
	})

	if err := database.Load(args.DBArgs); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	kongCtx.BindTo(ctx, (*context.Context)(nil))
	kongCtx.Bind(kongCtx)

	kongCtx.FatalIfErrorf(kongCtx.Run(args.Environment == "production"))
}
