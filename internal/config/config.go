package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       App       `mapstructure:"app"`
	Server    Server    `mapstructure:"server"`
	Log       Log       `mapstructure:"log"`
	MySQL     MySQL     `mapstructure:"mysql"`
	Postgres  Postgres  `mapstructure:"postgres"`
	SQLite    SQLite    `mapstructure:"sqlite"`
	Redis     Redis     `mapstructure:"redis"`
	JWT       JWT       `mapstructure:"jwt"`
	Cors      Cors      `mapstructure:"cors"`
	RateLimit RateLimit `mapstructure:"rate_limit"`
	Pprof     Pprof     `mapstructure:"pprof"`
	Metrics   Metrics   `mapstructure:"metrics"`
	Security  Security  `mapstructure:"security"`
}

type App struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	Debug    bool   `mapstructure:"debug"`
	Timezone string `mapstructure:"timezone"`
}

type Server struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes string        `mapstructure:"max_header_bytes"`
	Salt           string        `mapstructure:"salt"`
}

type Log struct {
	Level  string  `mapstructure:"level"`
	Format string  `mapstructure:"format"`
	Output string  `mapstructure:"output"`
	File   LogFile `mapstructure:"file"`
}

type LogFile struct {
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type MySQL struct {
	Enable          bool          `mapstructure:"enable"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Redis struct {
	Enable       bool          `mapstructure:"enable"`
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type JWT struct {
	Secret        string        `mapstructure:"secret"`
	Issuer        string        `mapstructure:"issuer"`
	Expire        time.Duration `mapstructure:"expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
}

type Cors struct {
	Enable           bool          `mapstructure:"enable"`
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type RateLimit struct {
	Enable bool `mapstructure:"enable"`
	Rps    int  `mapstructure:"rps"`
	Burst  int  `mapstructure:"burst"`
}

type Pprof struct {
	Enable bool   `mapstructure:"enable"`
	Addr   string `mapstructure:"addr"`
}

type Metrics struct {
	Enable bool   `mapstructure:"enable"`
	Path   string `mapstructure:"path"`
}

type Security struct {
	TrustedProxies []string `mapstructure:"trusted_proxies"`
	HideBanner     bool     `mapstructure:"hide_banner"`
}

type Postgres struct {
	Enable          bool          `mapstructure:"enable"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type SQLite struct {
	Enable          bool          `mapstructure:"enable"`
	File            string        `mapstructure:"file"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

var cfgFile string

var CFG Config

var Debug bool

func SetConfigFile(path string) {
	cfgFile = path
}

func MustNewConfig() {
	// load config file
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	replace := strings.NewReplacer(".", "_") // 替换点为下划线
	viper.SetEnvKeyReplacer(replace)         // 设置环境变量的替换器
	viper.AutomaticEnv()

	// unmarshal config
	if err := viper.Unmarshal(&CFG); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
