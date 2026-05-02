package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/app"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/joho/godotenv"

	_ "github.com/go-park-mail-ru/2026_1_Clac_Clac/docs"
)

// @title			NeXuS
// @description	Наше API
// @host			clac-clac.mooo.com
// @BasePath		/api
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

	a, err := app.NewApp(&conf)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	a.Run()
}
