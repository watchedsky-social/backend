package cli

type Args struct {
	Host     string `env:"DB_HOST" default:"pg.lab.verysmart.house" help:"host"`
	Username string `env:"DB_USER" default:"watchedsky-social" help:"user"`
	Password string `env:"DB_PASSWORD" help:"db password"`
	DB       string `env:"DB_NAME" default:"watchedsky-social"`
}
