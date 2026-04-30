package services

import (
	"fmt"
	"strings"

	"github.com/Alexander272/HasBot/internal/config"
	"github.com/Alexander272/HasBot/pkg/homeassistant"
	"github.com/Alexander272/HasBot/pkg/logger"
	"github.com/Alexander272/HasBot/pkg/mattermost"
)

type Service struct {
	haClient   *homeassistant.Client
	mostClient *mattermost.Client
	config     *config.Config
}

func (s *Service) Config() *config.Config {
	return s.config
}

func NewService(haClient *homeassistant.Client, mostClient *mattermost.Client, cfg *config.Config) *Service {
	return &Service{
		haClient:   haClient,
		mostClient: mostClient,
		config:     cfg,
	}
}

func (s *Service) HandleMessage(channelId string, msg string) error {
	logger.Debug("HandleMessage", logger.AnyAttr("msg", msg), logger.AnyAttr("channel", channelId))

	msg = strings.TrimSpace(msg)
	msgLower := strings.ToLower(msg)

	switch {
	case strings.HasPrefix(msgLower, "temp"),
		strings.HasPrefix(msgLower, "темп"),
		strings.HasPrefix(msgLower, "temperature"),
		strings.HasPrefix(msgLower, "температура"):
		return s.handleTemp(channelId, msg)
	case strings.HasPrefix(msgLower, "help"),
		strings.HasPrefix(msgLower, "man"),
		strings.HasPrefix(msgLower, "помощь"),
		strings.HasPrefix(msgLower, "мануал"):
		return s.handleHelp(channelId)
	default:
		return s.handleHelp(channelId)
	}
}

func (s *Service) handleTemp(channelId string, msg string) error {
	logger.Debug("handleTemp", logger.AnyAttr("msg", msg), logger.AnyAttr("channel", channelId))

	msgLower := strings.ToLower(msg)

	prefixes := []string{"temp", "темп", "temperature", "температура"}
	var cmdArgs []string
	for _, p := range prefixes {
		if strings.HasPrefix(msgLower, p) {
			cmdArgs = strings.Fields(strings.TrimPrefix(msgLower, p))
			break
		}
	}

	logger.Debug("handleTemp", logger.AnyAttr("args", cmdArgs))

	sensors := s.getSensorsForChannel(channelId)

	var filteredSensors []config.SensorConfig

	noArgs := len(cmdArgs) == 0

	if noArgs {
		logger.Debug("no args, returning all sensors")
		filteredSensors = sensors
	} else {
		arg := strings.Join(cmdArgs, " ")
		arg = strings.ToLower(arg)
		if arg == "all" || arg == "все" {
			filteredSensors = sensors
		} else {
			for _, sensor := range sensors {
				if strings.EqualFold(strings.ToLower(sensor.Name), arg) {
					filteredSensors = append(filteredSensors, sensor)
					break
				}
			}
			if len(filteredSensors) == 0 {
				return s.mostClient.SendMessage(channelId, "Датчик не найден. Используйте `temp all` для всех датчиков.")
			}
		}
	}

	if len(filteredSensors) == 0 {
		return s.mostClient.SendMessage(channelId, "Нет доступных датчиков.")
	}

	roomSensors := make(map[string][]config.SensorConfig)
	for _, sensor := range filteredSensors {
		room := sensor.Room
		if room == "" {
			room = "Другие"
		}
		roomSensors[room] = append(roomSensors[room], sensor)
	}

	var message string
	for room, roomSensorList := range roomSensors {
		message += fmt.Sprintf("**%s**\n", room)
		for _, sensor := range roomSensorList {
			state, err := s.haClient.GetState(sensor.EntityID)
			if err != nil {
				message += fmt.Sprintf("  • %s: ошибка\n", sensor.Name)
				continue
			}
			unit := ""
			if state.UnitOfMeasurement != "" {
				unit = " " + state.UnitOfMeasurement
			}
			message += fmt.Sprintf("  • %s: ___%s%s___\n", sensor.Name, state.State, unit)
		}
		message += "\n"
	}

	return s.mostClient.SendMessage(channelId, message)
}

func (s *Service) IsChannelAllowed(channelId string) bool {
	for _, ch := range s.config.Bot.Channels {
		if ch.ChannelId == channelId {
			return true
		}
	}
	return false
}

func (s *Service) getSensorsForChannel(channelId string) []config.SensorConfig {
	for _, ch := range s.config.Bot.Channels {
		if ch.ChannelId == channelId {
			return ch.Sensors
		}
	}
	return []config.SensorConfig{}
}

func (s *Service) handleHelp(channelId string) error {
	sensors := s.getSensorsForChannel(channelId)

	roomSensors := make(map[string][]config.SensorConfig)
	for _, sensor := range sensors {
		room := sensor.Room
		if room == "" {
			room = "Другие"
		}
		roomSensors[room] = append(roomSensors[room], sensor)
	}

	helpText := "Доступные команды:\n" +
		"• `temp` / `темп` - показать температуру всех датчиков\n" +
		"• `temp <имя>` / `темп <имя>` - показать конкретный датчик\n" +
		"• `help` / `помощь` - показать эту справку\n\n"

	for room, sensorList := range roomSensors {
		helpText += fmt.Sprintf("**%s**\n", room)
		for _, sensor := range sensorList {
			helpText += fmt.Sprintf("  • %s\n", sensor.Name)
		}
	}

	return s.mostClient.SendMessage(channelId, helpText)
}
