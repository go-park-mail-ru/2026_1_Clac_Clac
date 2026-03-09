package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
)

func main() {
	const configPath = "."

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
