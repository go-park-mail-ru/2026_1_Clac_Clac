package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/joho/godotenv"
)

// @title			NeXus
// @description	Наше API
// @host			clac-clac.mooo.com
// @BasePath		/api
func main() {
	const configPath = "."

	if err := godotenv.Load(); err != nil {
		log.Println("cannot find .env")
	}

	v, err := config.SetupViper(configPath)
	if err != nil {
		log.Fatalf("config.SetupViper: %v", err)
	}

	conf := config.DefaultConfig()
	if err := v.Unmarshal(&conf); err != nil {
		log.Fatalf("viper.Unmarshal: %v", err)
	}

	app := internal.NewApp(&conf)
	app.Run()
}
