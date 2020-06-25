package config

import (
	"fmt"
)

// Config configures the app
type Config struct {
	Facebook   Facebook
	Server   Server
	Database Database
	Api      Api
	Redis    Redis
	Ai       Ai
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
}

type Facebook struct {
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
