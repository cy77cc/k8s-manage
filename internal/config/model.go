package config

import "time"

type Config struct {
	App       App       `yaml:"app"`
	Server    Server    `yaml:"server"`
	Log       Log       `yaml:"log"`
	MySQL     MySQL     `yaml:"mysql"`
	Redis     Redis     `yaml:"redis"`
	JWT       JWT       `yaml:"jwt"`
	Cors      Cors      `yaml:"cors"`
	RateLimit RateLimit `yaml:"rate_limit"`
	Pprof     Pprof     `yaml:"pprof"`
	Metrics   Metrics   `yaml:"metrics"`
	Security  Security  `yaml:"security"`
}

type App struct {
	Name     string `yaml:"name"`
	Env      string `yaml:"env"`
	Debug    bool   `yaml:"debug"`
	Timezone string `yaml:"timezone"`
}

type Server struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes string        `yaml:"max_header_bytes"`
}

type Log struct {
	Level  string  `yaml:"level"`
	Format string  `yaml:"format"`
	Output string  `yaml:"output"`
	File   LogFile `yaml:"file"`
}

type LogFile struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

type MySQL struct {
	Enable          bool          `yaml:"enable"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	Charset         string        `yaml:"charset"`
	ParseTime       bool          `yaml:"parse_time"`
	Loc             string        `yaml:"loc"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type Redis struct {
	Enable       bool          `yaml:"enable"`
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type JWT struct {
	Secret        string        `yaml:"secret"`
	Issuer        string        `yaml:"issuer"`
	Expire        time.Duration `yaml:"expire"`
	RefreshExpire time.Duration `yaml:"refresh_expire"`
}

type Cors struct {
	Enable           bool          `yaml:"enable"`
	AllowOrigins     []string      `yaml:"allow_origins"`
	AllowMethods     []string      `yaml:"allow_methods"`
	AllowHeaders     []string      `yaml:"allow_headers"`
	ExposeHeaders    []string      `yaml:"expose_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

type RateLimit struct {
	Enable bool `yaml:"enable"`
	Rps    int  `yaml:"rps"`
	Burst  int  `yaml:"burst"`
}

type Pprof struct {
	Enable bool   `yaml:"enable"`
	Addr   string `yaml:"addr"`
}

type Metrics struct {
	Enable bool   `yaml:"enable"`
	Path   string `yaml:"path"`
}

type Security struct {
	TrustedProxies []string `yaml:"trusted_proxies"`
	HideBanner     bool     `yaml:"hide_banner"`
}
