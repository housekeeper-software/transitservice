package conf

import (
	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Opensips struct {
	SipServer  string `yaml:"sipServer"`
	StunServer string `yaml:"stunServer"`
	Domain     string `yaml:"domain"`
}
type Transit struct {
	Url  string `yaml:"url"`
	SUrl string `yaml:"surl"`
}

type Mysql struct {
	Url             string `yaml:"url"`
	Table           string `yaml:"table"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	ConnMaxLifeTime int    `yaml:"connMaxLifeTime"`
}
type Push struct {
	AppKey           string `yaml:"appKey"`
	AppSecret        string `yaml:"appSecret"`
	AppAccid         string `yaml:"appAccid"`
	SendAttachMsgUrl string `yaml:"sendAttachMsgUrl"`
	MsgTag           string `yaml:"msgTag"`
	Save             string `yaml:"save"`
}

type ServerConfig struct {
	Opensips Opensips `yaml:"opensips"`
	Transit  Transit  `yaml:"transit"`
	Mysql    Mysql    `yaml:"mysql"`
	Push     Push     `yaml:"push"`
}

func LoadServerConfig(file string) (*ServerConfig, error) {
	viper.SetConfigFile(file)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var c ServerConfig
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *ServerConfig) WriteConfigToFile(file string) error {
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, 0644)
}

func (c *ServerConfig) IsSupportPush() bool {
	return len(c.Push.SendAttachMsgUrl) > 0 &&
		len(c.Push.AppAccid) > 0 &&
		len(c.Push.AppKey) > 0 &&
		len(c.Push.AppSecret) > 0
}
