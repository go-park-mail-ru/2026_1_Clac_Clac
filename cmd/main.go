package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/subosito/gotenv"
)

func main() {
	const configPath = "."

	if err := gotenv.Load(); err != nil {
		log.Fatalf("gotenv.Load: %v", err)
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
