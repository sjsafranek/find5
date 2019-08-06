package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sjsafranek/find5/api"
	"github.com/sjsafranek/ligneous"
)

const (
	PROJECT                   string = "Find"
	VERSION                   string = "5.0.1"
	DEFAULT_HTTP_PORT         int    = 5555
	DEFAULT_DATABASE_ENGINE   string = "postgres"
	DEFAULT_DATABASE_DATABASE string = "finddb"
	DEFAULT_DATABASE_PASSWORD string = "dev"
	DEFAULT_DATABASE_USERNAME string = "finduser"
	DEFAULT_DATABASE_HOST     string = "localhost"
	DEFAULT_DATABASE_PORT     int64  = 5432
	DEFAULT_REDIS_PORT        int64  = 6379
	DEFAULT_REDIS_HOST        string = ""
)

var (
	HTTP_PORT         int    = DEFAULT_HTTP_PORT
	DATABASE_ENGINE   string = DEFAULT_DATABASE_ENGINE
	DATABASE_DATABASE string = DEFAULT_DATABASE_DATABASE
	DATABASE_PASSWORD string = DEFAULT_DATABASE_PASSWORD
	DATABASE_USERNAME string = DEFAULT_DATABASE_USERNAME
	DATABASE_HOST     string = DEFAULT_DATABASE_HOST
	DATABASE_PORT     int64  = DEFAULT_DATABASE_PORT
	REDIS_PORT        int64  = DEFAULT_REDIS_PORT
	REDIS_HOST        string = DEFAULT_REDIS_HOST
	REQUEST           string = ""
	logger                   = ligneous.AddLogger("server", "trace", "./log/find5")
	findapi           *api.Api
)

func init() {
	var print_version bool

	// flag.StringVar(&ACTION, "action", DEFAULT_ACTION, "Action")
	// flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Config file")
	// flag.BoolVar(&DEBUG, "debug", false, "debug mode")
	//
	flag.IntVar(&HTTP_PORT, "port", DEFAULT_HTTP_PORT, "Server port")
	flag.StringVar(&DATABASE_HOST, "dbhost", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&DATABASE_DATABASE, "dbname", DEFAULT_DATABASE_DATABASE, "database name")
	flag.StringVar(&DATABASE_PASSWORD, "dbpass", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&DATABASE_USERNAME, "dbuser", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&DATABASE_PORT, "dbport", DEFAULT_DATABASE_PORT, "Database port")
	flag.StringVar(&REDIS_HOST, "redishost", DEFAULT_REDIS_HOST, "Redis host")
	flag.StringVar(&REQUEST, "c", "", "Api query to execute")
	flag.Int64Var(&REDIS_PORT, "redisport", DEFAULT_REDIS_PORT, "Redis port")
	flag.BoolVar(&print_version, "V", false, "Print version and exit")
	flag.Parse()

	if print_version {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}

}

func main() {

	dbConnectionString := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable", DATABASE_ENGINE, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_HOST, DATABASE_PORT, DATABASE_DATABASE)
	redisAddr := fmt.Sprintf("%v:%v", REDIS_HOST, REDIS_PORT)
	findapi = api.New(dbConnectionString, redisAddr)

	if "" != REQUEST {
		request := api.Request{}
		request.Unmarshal(REQUEST)
		response, err := findapi.Do(&request)
		if nil != err {
			panic(err)
		}

		results, err := response.Marshal()
		if nil != err {
			panic(err)
		}
		fmt.Println(results)
		return
	}
}
