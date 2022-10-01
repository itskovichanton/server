package server

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/mitchellh/mapstructure"
)

// Loads core settings
type IConfigService interface {
	LoadConfig() (*Config, error)
}

type ConfigServiceImpl struct {
	IConfigService

	Config *core.Config
}

func (c *ConfigServiceImpl) LoadConfig() (*Config, error) {
	r := &Config{}
	return r, mapstructure.Decode(c.Config.Get("server"), &r.Server)
}

type Config struct {
	CoreConfig *core.Config
	Server     *Server
}

type Server struct {
	Port               int
	Http               *Http
	GrpcPort           int
	EnableThrottleMode bool
	EnableCORS         bool
	EnableGzip         bool
	DefaultLang        string
}

type Multipart struct {
	MaxRequestSizeBytes string
}

func (c Multipart) GetMaxRequestSizeBytes() (uint64, error) {
	return utils.ParseMemory(c.MaxRequestSizeBytes)
}

type Ssl struct {
	CertFile string
	KeyFile  string
	Enabled  bool
	Network  string
}

type Http struct {
	Multipart *Multipart
	Ssl       *Ssl
}
