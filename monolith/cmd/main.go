package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/app"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
	_ "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/docs"
	"github.com/joho/godotenv"
)

// @title			NeXus
// @description	Наше API
// @host			clac-clac.mooo.com
// @BasePath		/api
func main() {
	const configPath = "./monolith"

	if err := godotenv.Load("monolith/.env"); err != nil {
		log.Println("Файл .env не найден, надеемся на системные ENV переменные")
	}

	v, err := config.SetupViper(configPath)
	if err != nil {
		log.Fatalf("config.SetupViper: %v", err)
	}

	conf := config.DefaultConfig()
	if err := v.Unmarshal(&conf); err != nil {
		log.Fatalf("viper.Unmarshal: %v", err)
	}

	app := app.NewApp(&conf)
	app.Run()
}
