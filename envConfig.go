package main

type postgresConfig struct {
	Host    string `env:"PG_HOST,required"`
	Port    string `env:"PG_PORT" envDefault:"5432"`
	User    string `env:"PG_USER,required"`
	Pass    string `env:"PG_PASS,required"`
	DBName  string `env:"PG_DBNAME,required"`
	SSLMode string `env:"PG_SSLMODE" envDefault:"disable"`
}

type redisConfig struct {
	Addr     string `env:"R_ADDR,required"`
	Password string `env:"R_PASS,required"`
	DB       int    `env:"R_DB" envDefault:"0"`
}

type telegramConfig struct {
	TelegramToken string `env:"TG_TOKEN,required"`
}
