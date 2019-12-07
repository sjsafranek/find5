package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sjsafranek/find5/findapi/lib/api"
	"github.com/sjsafranek/find5/findapi/lib/clients/repl"
	"github.com/sjsafranek/find5/findapi/lib/clients/web"
	// "github.com/sjsafranek/logger"
)

const (
	PROJECT                   string = "Find"
	VERSION                   string = "5.0.6"
	DEFAULT_HTTP_PORT         int    = 8080
	DEFAULT_DATABASE_ENGINE   string = "postgres"
	DEFAULT_DATABASE_DATABASE string = "finddb"
	DEFAULT_DATABASE_PASSWORD string = "dev"
	DEFAULT_DATABASE_USERNAME string = "finduser"
	DEFAULT_DATABASE_HOST     string = "localhost"
	DEFAULT_DATABASE_PORT     int64  = 5432
	DEFAULT_REDIS_PORT        int64  = 6379
	DEFAULT_REDIS_HOST        string = ""
	DEFAULT_AI_HOST           string = "localhost"
	DEFAULT_AI_PORT           int64  = 7005
	// DEFAULT_LOGGING_DIRECTORY string = ""
	DEFAULT_CONFIG_FILE       string = "config.json"
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
	// LOGGING_DIRECTORY string = DEFAULT_LOGGING_DIRECTORY
	AI_HOST           string = DEFAULT_AI_HOST
	AI_PORT           int64  = DEFAULT_AI_PORT
	CONFIG_FILE       string = DEFAULT_CONFIG_FILE
	REQUEST           string = ""
	MODE              string = "web"
	// logger            ligneous.Log
	findapi           *api.Api
)

func init() {
	var printVersion bool

	// flag.StringVar(&ACTION, "action", DEFAULT_ACTION, "Action")
	// flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Config file")
	// flag.BoolVar(&VERBOSE, "verbose", false, "verbose messaging")

	flag.IntVar(&HTTP_PORT, "httpport", DEFAULT_HTTP_PORT, "Server port")
	flag.StringVar(&DATABASE_HOST, "dbhost", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&DATABASE_DATABASE, "dbname", DEFAULT_DATABASE_DATABASE, "database name")
	flag.StringVar(&DATABASE_PASSWORD, "dbpass", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&DATABASE_USERNAME, "dbuser", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&DATABASE_PORT, "dbport", DEFAULT_DATABASE_PORT, "Database port")
	flag.StringVar(&REDIS_HOST, "redishost", DEFAULT_REDIS_HOST, "Redis host")
	flag.Int64Var(&REDIS_PORT, "redisport", DEFAULT_REDIS_PORT, "Redis port")
	flag.StringVar(&AI_HOST, "aihost", DEFAULT_AI_HOST, "AI host")
	flag.Int64Var(&AI_PORT, "aiport", DEFAULT_AI_PORT, "AI port")
	// flag.StringVar(&LOGGING_DIRECTORY, "L", DEFAULT_LOGGING_DIRECTORY, "Logging directory")
	flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "config file")
	flag.StringVar(&REQUEST, "query", "", "Api query to execute")
	flag.BoolVar(&printVersion, "V", false, "Print version and exit")
	flag.Parse()

	// logger = ligneous.AddLogger("server", "trace", fmt.Sprintf("%v/find5", LOGGING_DIRECTORY))

	if printVersion {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}

	args := flag.Args()
	if 1 == len(args) {
		MODE = args[0]
	}

}

func main() {
	// api.SetLoggingDirectory(LOGGING_DIRECTORY)
	// web.SetLoggingDirectory(LOGGING_DIRECTORY)
	// repl.SetLoggingDirectory(LOGGING_DIRECTORY)

	dbConnectionString := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable", DATABASE_ENGINE, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_HOST, DATABASE_PORT, DATABASE_DATABASE)
	redisAddr := fmt.Sprintf("%v:%v", REDIS_HOST, REDIS_PORT)
	aiConnStr := fmt.Sprintf("%v:%v", AI_HOST, AI_PORT)
	findapi = api.New(dbConnectionString, aiConnStr, redisAddr)

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

	switch MODE {
	case "repl":
		repl.New(findapi).Run()
		break
	case "web":
		web.New(findapi).Run(HTTP_PORT)
		break
	default:
		panic(errors.New("api client not found"))
	}

}
