package common

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigParser struct {
	v *viper.Viper
	l *logrus.Logger
}

func NewConfigParser(l *logrus.Logger, prefix string) *ConfigParser {
	v := viper.New()
	v.SetConfigType("toml")

	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()

	return &ConfigParser{
		v: v,
		l: l,
	}
}

func (p *ConfigParser) ReadFile(path string) {
	p.v.SetConfigFile(path)
	err := p.v.ReadInConfig()
	if err != nil {
		p.l.Info("config file not found, using defaults")
	}
}

func (p *ConfigParser) GetString(key string, fallback string) string {
	p.v.SetDefault(key, fallback)
	return p.v.GetString(key)
}

func (p *ConfigParser) GetInt64(key string, fallback int64) int64 {
	p.v.SetDefault(key, fallback)
	return p.v.GetInt64(key)
}
