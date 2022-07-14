package application

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

const DefaultConfigDir = "."

type Configurator struct {
	viper *viper.Viper
}

type Clouds struct {
	AWSTenants   []AWSTenant   `mapstructure:"aws"`
	AzureTenants []AzureTenant `mapstructure:"azure"`
	GCPTenants   []GCPTenant   `mapstructure:"gcp"`
}

type AWSTenant struct {
	Name            string `mapstructure:"name"`
	MasterAccountID string `mapstructure:"master_account_id"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
}

type AzureTenant struct{}

type GCPTenant struct{}

type OktaConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Issuer   string `mapstructure:"issuer"`
	ClientID string `mapstructure:"client_id"`
}

type FrontendConfig struct {
	URL string `mapstructure:"url"`
}

type DBConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Port     int64  `mapstructure:"port"`
	Host     string `mapstructure:"host"`
}

type NatsConfig struct {
	URL string `mapstructure:"url"`
}

func (db *DBConfig) GetDSN() string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", db.User, db.Password, db.Host, db.Port, db.Database)
	return dsn
}

type ServerConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type Config struct {
	Okta         OktaConfig     `mapstructure:"okta"`
	DB           DBConfig       `mapstructure:"db"`
	Frontend     FrontendConfig `mapstructure:"frontend"`
	Clouds       Clouds         `mapstructure:"clouds"`
	Nats         NatsConfig     `mapstructure:"nats"`
	ServerConfig ServerConfig   `mapstructure:"server"`
}

func NewConfigurator(configDir string) Configurator {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	return Configurator{
		viper: v,
	}
}

func (c *Configurator) Parse() (Config, error) {
	err := c.viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	conf := defaultConfig()
	err = c.viper.Unmarshal(&conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func defaultConfig() Config {
	return Config{
		Clouds: Clouds{
			AWSTenants:   []AWSTenant{},
			GCPTenants:   []GCPTenant{},
			AzureTenants: []AzureTenant{},
		},
		Nats: NatsConfig{
			URL: nats.DefaultURL,
		},
		Okta: OktaConfig{
			Enabled: false,
		},
		ServerConfig: ServerConfig{
			BaseURL: "http://localhost:8080",
		},
	}
}
