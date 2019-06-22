package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Parser struct {
	v *viper.Viper
	l *logrus.Logger
}

func NewParser(l *logrus.Logger, prefix string) *Parser {
	v := viper.New()
	v.SetConfigType("toml")

	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()

	return &Parser{
		v: v,
		l: l,
	}
}

func (p *Parser) ReadFile(path string) {
	p.v.SetConfigFile(path)
	err := p.v.ReadInConfig()
	if err != nil {
		p.l.Info("config file not found, using defaults")
	}
}

func (p *Parser) GetString(key string, fallback string) string {
	p.v.SetDefault(key, fallback)
	return p.v.GetString(key)
}

func (p *Parser) GetInt64(key string, fallback int64) int64 {
	p.v.SetDefault(key, fallback)
	return p.v.GetInt64(key)
}
