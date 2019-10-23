package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/motoki317/github-webhook/webhook"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Server successfully started!")
	})

	e.POST("/webhook", webhook.MakeWebhookHandler())

	port := os.Getenv("WEBHOOK_PORT")
	if port == "" {
		log.Println("Env-var WEBHOOK_PORT empty! Setting default port to 8090.")
		port = "8090"
	}

	log.Fatal(e.Start(":" + port))
}
