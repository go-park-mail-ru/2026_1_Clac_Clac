package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/app"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	const configPath = "."

	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	v, err := config.SetupViper(configPath)
	if err != nil {
		log.Fatalf("config.SetupViper: %v", err)
	}

	conf := config.DefaultConfig()
	if err := v.Unmarshal(&conf); err != nil {
		log.Fatalf("viper.Unmarshal: %v", err)
	}

	app, err := app.NewApp(&conf)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	app.Run()
}
