package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alexander272/HasBot/internal/config"
	"github.com/Alexander272/HasBot/internal/services"
	"github.com/Alexander272/HasBot/internal/transport/socket"
	"github.com/Alexander272/HasBot/pkg/homeassistant"
	"github.com/Alexander272/HasBot/pkg/logger"
	"github.com/Alexander272/HasBot/pkg/mattermost"
	"github.com/subosito/gotenv"
)

func main() {
	if os.Getenv("APP_ENV") == "" {
		if err := gotenv.Load(".env"); err != nil {
			log.Fatalf("failed to load env variables. error: %s", err.Error())
		}
	}

	conf, err := config.Init("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to init configs. error: %s", err.Error())
	}
	logger.NewLogger(logger.WithLevel(conf.LogLevel), logger.WithAddSource(conf.LogSource))

	mattermostConf := mattermost.Config{
		ServerLink: conf.Bot.Server,
		Token:      conf.Bot.Token,
	}
	mostClient := mattermost.NewClient(mattermostConf)

	_, _, err = mostClient.Http.GetPing()
	if err != nil {
		log.Fatalf("failed to ping mattermost. error: %s", err.Error())
	}

	bot, _, err := mostClient.Http.GetMe("")
	if err != nil {
		log.Fatalf("failed to get bot data. error: %s", err.Error())
	}
	logger.Debug("me", logger.AnyAttr("bot", bot))

	haClient := homeassistant.NewClient(homeassistant.Config{
		Url:   conf.HomeAssistant.Url,
		Token: conf.HomeAssistant.Token,
	})

	service := services.NewService(haClient, mostClient, conf)

	if !mostClient.Connect() {
		log.Fatalf("failed to connect to mattermost websocket")
	}

	socHandler := socket.NewHandler(mostClient.Socket, bot, service)

	go func() {
		socHandler.Listen()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	socHandler.Close()
}
