package config

type PGConfig struct {
	Port    int     `env:"PG_PORT" envDefault:"5432"`
	Host    string  `env:"DATABASE_HOST,required"`
	User 	string  `env:"PG_USER,required"`
	Pass	string 	`env:"PG_PASS,required"`
	Name	string 	`env:"DB_NAME" envDefault:"postgres"`
	Ssl		string	`env:"PG_USE_SSL" envDefault:"disable"`
}

type MGConfig struct {
    Host     string `env:"MG_HOST" envDefault:"localhost"`
    Port     int    `env:"MG_PORT" envDefault:"7687"`
    User     string `env:"MG_USER" envDefault:""`
    Pass     string `env:"MG_PASS" envDefault:""`
}