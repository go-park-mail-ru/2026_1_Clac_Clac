package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/app"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	const configPath = "./facade"

	if err := godotenv.Load("facade/.env"); err != nil {
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
