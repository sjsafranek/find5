package repl

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/sjsafranek/find5/api"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("RUN"),
	readline.PcItem("BYE"),
	readline.PcItem("EXIT"),
	readline.PcItem("HELP"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

type Client struct {
	api *api.Api
}

func (self *Client) Run() {

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m[find5]#\033[0m ",
		HistoryFile:         "history.find5",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		command := strings.ToLower(parts[0])

		// testing
		setPasswordCfg := l.GenPasswordConfig()
		setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
			l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
			l.Refresh()
			return nil, 0, false
		})
		//.end

		switch {

		// case "run" == command:
		case strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}"):
			request := api.Request{}
			request.Unmarshal(line)
			response, _ := self.api.Do(&request)
			results, _ := response.Marshal()
			fmt.Println(results)

		case "bye" == command:
			goto exit

		case "exit" == command:
			goto exit

		case "quit" == command:
			goto exit

		case line == "":
		default:
			// log.Println("you said:", strconv.Quote(line))
		}
	}
exit:
}

func New(findapi *api.Api) *Client {

	return &Client{api: findapi}

}
