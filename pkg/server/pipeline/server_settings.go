package pipeline

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server/entities"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type IServerSettingsProviderService interface {
	GetSettings() *GlobalSettings
	GetSecurity() *Security
}

func (c *ServerSettingsProviderServiceImpl) GetSettings() *GlobalSettings {
	return c.Settings
}

func (c *ServerSettingsProviderServiceImpl) GetSecurity() *Security {
	return c.Security
}

type ServerSettingsProviderServiceImpl struct {
	IServerSettingsProviderService

	Settings *GlobalSettings
	Security *Security

	Config *core.Config
}

func (c *ServerSettingsProviderServiceImpl) Reload() error {
	err := c.reloadSettings()
	if err != nil {
		return err
	}
	err = c.reloadSecurity()
	if err != nil {
		return nil
	}
	return nil
}

func (c *ServerSettingsProviderServiceImpl) reloadSecurity() error {

	f, err := c.Config.GetSecurityFile()
	if err != nil {
		return err
	}
	err = utils.UnmarshalYaml(f.Name(), &c.Security)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServerSettingsProviderServiceImpl) reloadSettings() error {
	f, err := c.Config.GetSettingsFile()
	if err != nil {
		return err
	}
	fn, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return err
	}

	var ss GlobalSettings
	err = yaml.Unmarshal(fn, &ss)
	c.Settings = &ss
	return nil
}

type GlobalSettings struct {
	Version      *entities.Version
	UpdateAppUrl string
}

type Security struct {
	Profiles map[string]*Profile
	Actions  map[string][]string
}

type Profile struct {
	Ips     []string
	Profile string
	Denied  bool
}
