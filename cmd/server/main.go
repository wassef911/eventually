package main

import (
	"flag"
	"log"

	"github.com/wassef911/eventually/internal/api"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

func main() {
	flag.Parse()

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(config.Logger)
	appLogger.InitLogger()
	appLogger.WithName(config.ServiceName)
	appLogger.Fatal(api.New(config, appLogger).Run())
}
