package config

import (
	"fmt"
)

// Config configures the app
type Config struct {
	OAuth2 OAuth2
	Server   Server
	Database Database
	Api      Api
	Redis    Redis
	Ai       Ai
}

type OAuth2 struct {
	Facebook SocialOAuth2
	Google SocialOAuth2
	GitHub SocialOAuth2
}

func (self *OAuth2) HasFacebook() bool {
	return "" != self.Facebook.ClientID && "" != self.Facebook.ClientSecret
}

func (self *OAuth2) HasGoogle() bool {
	return "" != self.Google.ClientID && "" != self.Google.ClientSecret
}

func (self *OAuth2) HasGitHub() bool {
	return "" != self.GitHub.ClientID && "" != self.GitHub.ClientSecret
}

type Api struct {
	PublicMethods []string
}

func (self *Api) IsPublicMethod(method string) bool {
	for _, publicMethod := range self.PublicMethods {
		if publicMethod == method {
			return true
		}
	}
	return false
}

type Server struct {
	HttpPort int
	HttpHost string
	HttpProtocol string
	HttpDomain string
}

func (self *Server) GetURLString() string {
	if "" == self.HttpDomain {
		return fmt.Sprintf("%v://%v:%v", self.HttpProtocol, self.HttpHost, self.HttpPort)
	}
	return fmt.Sprintf("%v://%v", self.HttpProtocol, self.HttpDomain)
}

type SocialOAuth2 struct {
	ClientID     string
	ClientSecret string
}

type Redis struct {
	Host string
	Port int64
}

func (self *Redis) GetConnectionString() string {
	return fmt.Sprintf("%v:%v", self.Host, self.Port)
}

type Ai struct {
	Host string
	Port int64
}

func (self *Ai) GetConnectionString() string {
	return fmt.Sprintf("%v:%v", self.Host, self.Port)
}

type Database struct {
	DatabaseEngine string
	DatabaseName   string
	DatabasePass   string
	DatabaseUser   string
	DatabaseHost   string
	DatabasePort   int64
}

func (self *Database) GetDatabaseConnection() string {
	return fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable",
		self.DatabaseEngine,
		self.DatabaseUser,
		self.DatabasePass,
		self.DatabaseHost,
		self.DatabasePort,
		self.DatabaseName)
}
