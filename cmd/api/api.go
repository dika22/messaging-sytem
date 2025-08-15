package cmd

import (
	"context"
	"fmt"
	"log"
	"multi-tenant-service/cmd/middleware"
	"multi-tenant-service/internal/tenant/delivery"
	"multi-tenant-service/internal/tenant/usecase"
	"multi-tenant-service/metrics"
	"multi-tenant-service/package/config"
	"multi-tenant-service/package/logger"
	"net/http"
	"os"
	"time"

	"os/signal"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/urfave/cli/v2"

	_ "multi-tenant-service/docs"

	um "multi-tenant-service/internal/message/usecase"

	deliMessage "multi-tenant-service/internal/message/delivery"
)

const CmdServeHTTP = "serve-http"

type HTTP struct {
	usecase usecase.ITenantUsecase
	um      um.IMessageUsecase
	cfg     *config.Config
}

func (h HTTP) ServeAPI(c *cli.Context) error {
	if err := logger.SetLogger(); err != nil {
		log.Printf("error logger %v", err)
	}
	// Register metrics
	metrics.Register()

	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	
	e.GET("/health-check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "ok!")
	})


	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "pong")
	})

	// Prometheus metrics endpoint
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	tenantAPI := e.Group("api/v1")

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		// AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	

	tenantAPI.Use(middleware.LoggerMiddleware)
	tenantAPI.Use(middleware.MonitoringMiddleware)

	delivery.NewTenantHTTPHandler(tenantAPI, h.usecase)
	deliMessage.NewMessageHTTPHandler(tenantAPI, h.um)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf(":%v", h.cfg.Server.Port)); err != nil {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	return nil
}

func ServeAPI(usecase usecase.ITenantUsecase, um um.IMessageUsecase, cfg *config.Config) []*cli.Command {
	h := &HTTP{usecase: usecase, um: um, cfg: cfg}
	return []*cli.Command{
		{
			Name:   CmdServeHTTP,
			Usage:  "Serve multi-tenant service",
			Action: h.ServeAPI,
		},
	}
}
