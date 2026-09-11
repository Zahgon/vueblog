// Package config mirrors com.markerhub.config plus the application.yml binding
// that Spring Boot performs via @ConfigurationProperties.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Properties mirrors the settings vueblog reads from application.yml and
// application-default.yml.
type Properties struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Markerhub struct {
		Jwt struct {
			Secret string `yaml:"secret"`
			Expire int64  `yaml:"expire"`
			Header string `yaml:"header"`
		} `yaml:"jwt"`
	} `yaml:"markerhub"`

	Spring struct {
		Datasource struct {
			URL      string `yaml:"url"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"datasource"`
	} `yaml:"spring"`

	ShiroRedis struct {
		Enabled      bool `yaml:"enabled"`
		RedisManager struct {
			Host string `yaml:"host"`
		} `yaml:"redis-manager"`
	} `yaml:"shiro-redis"`
}

// Defaults mirrors the committed application.yml values, so the service starts
// with identical settings when no file is supplied.
func Defaults() *Properties {
	p := &Properties{}
	p.Server.Port = 8081
	p.Markerhub.Jwt.Secret = "f4e2e52034348f86b67cde581c0f9eb5"
	p.Markerhub.Jwt.Expire = 604800
	p.Markerhub.Jwt.Header = "Authorization"
	p.Spring.Datasource.URL = "root:admin@tcp(localhost:3306)/vueblog?charset=utf8&parseTime=true&loc=Local"
	p.Spring.Datasource.Username = "root"
	p.Spring.Datasource.Password = "admin"
	p.ShiroRedis.Enabled = true
	p.ShiroRedis.RedisManager.Host = "127.0.0.1:6379"
	return p
}

// Load reads a YAML file over the defaults.
func Load(path string) (*Properties, error) {
	p := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, p); err != nil {
		return nil, err
	}
	return p, nil
}
