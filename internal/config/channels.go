package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/Alexander272/HasBot/internal/models"
	"gopkg.in/yaml.v3"
)

type channelsFile struct {
	Channels []models.ChannelConfig `yaml:"channels"`
}

type ChannelsStore struct {
	mu       sync.RWMutex
	channels []models.ChannelConfig
	filePath string
}

func NewChannelsStore(filePath string) (*ChannelsStore, error) {
	s := &ChannelsStore{filePath: filePath}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.channels = []models.ChannelConfig{}
			return s, nil
		}
		return nil, fmt.Errorf("failed to read channels file: %w", err)
	}

	var cf channelsFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channels: %w", err)
	}

	if cf.Channels == nil {
		cf.Channels = []models.ChannelConfig{}
	}

	s.channels = cf.Channels
	return s, nil
}

func (s *ChannelsStore) save() error {
	cf := channelsFile{Channels: s.channels}
	data, err := yaml.Marshal(&cf)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write channels file: %w", err)
	}
	return nil
}

func (s *ChannelsStore) GetAll() []models.ChannelConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.ChannelConfig, len(s.channels))
	copy(result, s.channels)
	return result
}

func (s *ChannelsStore) GetByChannelId(channelId string) (models.ChannelConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.channels {
		if ch.ChannelId == channelId {
			return ch, true
		}
	}
	return models.ChannelConfig{}, false
}

func (s *ChannelsStore) IsAllowed(channelId string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.channels {
		if ch.ChannelId == channelId {
			return true
		}
	}
	return false
}

func (s *ChannelsStore) GetSensors(channelId string) []models.SensorConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.channels {
		if ch.ChannelId == channelId {
			return ch.Sensors
		}
	}
	return []models.SensorConfig{}
}

func (s *ChannelsStore) ReplaceAll(channels []models.ChannelConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if channels == nil {
		channels = []models.ChannelConfig{}
	}

	s.channels = channels
	return s.save()
}

func (s *ChannelsStore) Delete(channelId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, ch := range s.channels {
		if ch.ChannelId == channelId {
			s.channels = append(s.channels[:i], s.channels[i+1:]...)
			return s.save()
		}
	}

	return fmt.Errorf("channel %s not found", channelId)
}
