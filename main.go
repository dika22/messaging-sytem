package main

import (
	"log"
	"multi-tenant-service/cmd/migrate"
	rm "multi-tenant-service/internal/message/repository"
	um "multi-tenant-service/internal/message/usecase"
	"multi-tenant-service/internal/tenant/repository"
	"multi-tenant-service/internal/tenant/usecase"
	"multi-tenant-service/package/config"
	"multi-tenant-service/package/connection/database"
	rabbitmq "multi-tenant-service/package/rabbit-mq"
	"os"

	api "multi-tenant-service/cmd/api"

	"github.com/urfave/cli/v2"
)

func main()  {
	// Load configuration
	cfg, err := config.Load("package/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := database.Connect(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		
	}

	mqClient, err := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Println("ERROR INIT RABBITMQ", err)
	}

	tenantRepo := repository.NewTenantRepository(dbConn)
	messageRepo := rm.NewMessageRepository(dbConn)

	messageUsecase := um.NewMessageUsecase(messageRepo, tenantRepo, mqClient)
	tenantUsecase := usecase.NewTenantUsecase(tenantRepo, messageRepo, mqClient)

	cmds := []*cli.Command{}
	cmds = append(cmds, api.ServeAPI(tenantUsecase, messageUsecase, cfg)...)
	cmds = append(cmds, migrate.NewMigrate(cfg)...)

	app := &cli.App{
		Name:     "messaging-system",
		Commands: cmds,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}