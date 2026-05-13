package models

type ChannelConfig struct {
	ChannelId string         `yaml:"channel_id" json:"channel_id"`
	Sensors   []SensorConfig `yaml:"sensors" json:"sensors"`
}

type SensorConfig struct {
	Name     string `yaml:"name" json:"name"`
	EntityID string `yaml:"entity_id" json:"entity_id"`
	Room     string `yaml:"room" json:"room"`
}
