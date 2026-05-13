package services

import (
	"fmt"
	"strings"

	"github.com/Alexander272/HasBot/internal/config"
	"github.com/Alexander272/HasBot/internal/models"
	"github.com/Alexander272/HasBot/pkg/homeassistant"
	"github.com/Alexander272/HasBot/pkg/logger"
	"github.com/Alexander272/HasBot/pkg/mattermost"
)

type Service struct {
	haClient   *homeassistant.Client
	mostClient *mattermost.Client
	config     *config.Config
	channels   *config.ChannelsStore
}

func (s *Service) Config() *config.Config {
	return s.config
}

func NewService(haClient *homeassistant.Client, mostClient *mattermost.Client, cfg *config.Config, ch *config.ChannelsStore) *Service {
	return &Service{
		haClient:   haClient,
		mostClient: mostClient,
		config:     cfg,
		channels:   ch,
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

	var filteredSensors []models.SensorConfig

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

	roomSensors := make(map[string][]models.SensorConfig)
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
	return s.channels.IsAllowed(channelId)
}

func (s *Service) getSensorsForChannel(channelId string) []models.SensorConfig {
	return s.channels.GetSensors(channelId)
}

func (s *Service) GetChannels() []models.ChannelConfig {
	return s.channels.GetAll()
}

func (s *Service) ReplaceChannels(channels []models.ChannelConfig) error {
	return s.channels.ReplaceAll(channels)
}

func (s *Service) DeleteChannel(channelId string) error {
	return s.channels.Delete(channelId)
}

func (s *Service) handleHelp(channelId string) error {
	sensors := s.getSensorsForChannel(channelId)

	roomSensors := make(map[string][]models.SensorConfig)
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
