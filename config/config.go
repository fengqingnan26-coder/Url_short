package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Server    ServerConfig    `mapstructure:"server"`
	App       AppConfig       `mapstructure:"app"`
	ShortCode ShortCodeConfig `mapstructure:"shortcode"`
	Filter    FilterConfig    `mapstructure:"filter"`
	Logger    LogConfig       `mapstructure:"logger"`
	Email     EmailConfig     `mapstructure:"email"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RandNum   RandNumConfig   `mapstructure:"rand_num"`
}

// var Cfg *Config

func NewFromFile(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)
	viper.SetEnvPrefix("URL_SHORTENER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"ssl_mode"`
	Charset      string `mapstructure:"charset"`    // MySQL 专用
	ParseTime    bool   `mapstructure:"parse_time"` // MySQL 专用
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (d DatabaseConfig) DNS() string {
	// encodedPassword := url.QueryEscape(d.Password)
	switch strings.ToLower(d.Driver) {
	case "mysql", "mariadb":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&tls=%s",
			d.User, d.Password, d.Host, d.Port, d.DBName, d.getTLSMode())
	default:
		log.Fatal("not support db")
		return ""
	}
}

func (d DatabaseConfig) getTLSMode() string {
	switch strings.ToLower(d.SSLMode) {
	case "require", "verify-ca", "verify-identity":
		return "true"
	case "disable":
		return "false"
	default:
		return "skip-verify"
	}
}

type RedisConfig struct {
	Address           string        `mapstructure:"address"`
	Password          string        `mapstructure:"password"`
	DB                int           `mapstructure:"db"`
	BloomFilterName   string        `mapstructure:"bloom_filter_name"`
	BloomErrorRate    float64       `mapstructure:"bloom_error_rate"`
	BloomCapacity     uint          `mapstructure:"bloom_capacity"`
	CacheTTL          time.Duration `mapstructure:"cache_ttl"`
	URLDuration       time.Duration `mapstructure:"url_duration"`
	EmailCodeDuration time.Duration `mapstructure:"email_code_duration"`
}

type ServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
}

type AppConfig struct {
	BaseURL          string        `mapstructure:"base_url"`
	DefaultDuration  time.Duration `mapstructure:"default_duration"`
	CleanupInterval  time.Duration `mapstructure:"cleanup_interval"`
	SyncViewDuration time.Duration `mapstructure:"sync_view_interval"`
}

type ShortCodeConfig struct {
	Length int `mapstructure:"length"`
}

type FilterConfig struct {
	Capacity  uint    `mapstructure:"capacity"`
	ErrorRate float64 `mapstructure:"error_rate"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type RandNumConfig struct {
	Length int `mapstructure:"length"`
}

type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	Duration time.Duration `mapstructure:"duration"`
}

type EmailConfig struct {
	Password    string `mapstructure:"password"`
	Username    string `mapstructure:"username"`
	HostAddress string `mapstructure:"host_address"`
	HostPort    string `mapstructure:"host_port"`
	Subject     string `mapstructure:"subject"`
	TestMail    string `mapstructure:"test_mail"`
}
