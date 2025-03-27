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

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewAppLogger(cfg.Logger)
	appLogger.InitLogger()
	appLogger.WithName(cfg.ServiceName)
	appLogger.Fatal(api.New(cfg, appLogger).Run())
}
