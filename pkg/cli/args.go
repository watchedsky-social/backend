package cli

import (
	"github.com/alecthomas/kong"
	"github.com/watchedsky-social/backend/pkg/cli/commands"
)

type DBArgs struct {
	Host     string `env:"DB_HOST" default:"pg.lab.verysmart.house" help:"host"`
	Username string `env:"DB_USER" default:"watchedsky-social" help:"user"`
	Password string `env:"DB_PASSWORD" help:"db password"`
	DB       string `env:"DB_NAME" default:"watchedsky-social"`
}

type ServerArgs struct {
	DBArgs
	Environment string                `short:"e" env:"ENV" enum:"dev,production" default:"dev" help:"The running environment"`
	Version     kong.VersionFlag      `short:"v" help:"Display this app's version and exit"`
	Server      *commands.HTTPCommand `cmd:"" default:"withargs" help:"Run the API server and UI"`
}
