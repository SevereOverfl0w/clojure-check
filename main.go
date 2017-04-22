package main

import (
	"fmt"
	"github.com/google/uuid"
	bencode "github.com/jackpal/bencode-go"
	"gopkg.in/edn.v1"
	"io/ioutil"
	"net"
	"os"
	"flag"
	// "os/signal"
	// "syscall"
	"regexp"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Response struct {
	Ex      string
	Value   string
	Status  []string
	Id      string
	Session string
	Out     string
}

type EastwoodArgs struct {
	Namespaces []edn.Symbol `edn:",omitempty"`
}

func printmsgid(conn net.Conn, msgid string) {
	for {
		result := Response{}
		err := bencode.Unmarshal(conn, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		if result.Id != msgid {
			fmt.Println("Skipping this message id")
			continue
		}
		if result.Ex != "" {
			fmt.Println(result.Ex)
		}

		if result.Out != "" {
			matched, _ := regexp.MatchString(".+:.+\n", result.Out)
			xmatch, _ := regexp.MatchString("==.+\n", result.Out)
			if matched && !xmatch {
				fmt.Print(result.Out)
			}
		}

		// if result.Value != "" {
		//     fmt.Println("value: ")
		//     fmt.Println(result.Value)
		// }

		if len(result.Status) > 0 {
			return
			// if stringInSlice("done", result.Status) {
			// return
			// }
		}
	}
}

func eastwood(args []string, conn net.Conn) {
	ns := make([]edn.Symbol, len(args))

	for index, element := range args {
		ns[index] = edn.Symbol(element)
	}

	x := EastwoodArgs{ns}
	b, err := edn.Marshal(x)
	code := fmt.Sprintf(`(do (require 'eastwood.lint) (eastwood.lint/eastwood '%v))`, string(b))

	msguuid, _ := uuid.NewRandom()
	msgid := msguuid.String()

	instruction := map[string]interface{}{
		"op":   "eval",
		"code": code,
		"id":   msgid,
	}
	err = bencode.Marshal(conn, instruction)

	if err != nil {
		fmt.Println(err)
		return
	}

	printmsgid(conn, msgid)
}

func kibit(input string, file bool, conn net.Conn) {
	var reader string

	reporter := `(fn [{:keys [file line expr alt]}] (printf "%s:%s:0: Consider using: %s Instead of %s\n" file line (pr-str alt) (pr-str expr)))`

	if file {
		escapedFileBin, err := edn.Marshal(input)

		if err != nil {
			fmt.Println(err)
			return
		}

		escapedFile := string(escapedFileBin)

		reader = fmt.Sprintf(`(kibit.check/check-file (java.io.File. %v) :reporter %v)`, escapedFile, reporter)
	} else {
		reader = fmt.Sprintf(`(run! %v (kibit.check/check-reader (java.io.StringReader. %v)))`, reporter, input)
	}

	code := fmt.Sprintf(`(do (require 'kibit.check) %v)`, reader)

	msguuid, _ := uuid.NewRandom()
	msgid := msguuid.String()

	instruction := map[string]interface{}{
		"op":   "eval",
		"code": code,
		"id":   msgid,
	}
	err := bencode.Marshal(conn, instruction)

	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		result := Response{}
		err := bencode.Unmarshal(conn, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		if result.Id != msgid {
			continue
		}
		if result.Ex != "" {
			fmt.Println(result.Ex)
		}

		if result.Out != "" {
			fmt.Print(result.Out)
		}

		if len(result.Status) > 0 {
			return
		}
	}
}

type namespaceFlags []string

func (i *namespaceFlags) String() string {
	return "my string representation"
}

func (i *namespaceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var namespaces namespaceFlags
	var file string
	var nrepl string
	flag.Var(&namespaces, "namespace", "Namespace to lint. Can be repeated multiple times")
	flag.StringVar(&file, "file", "-", "File to lint via kibit. If - will be read from stdin. Default is -")
	flag.StringVar(&nrepl, "nrepl", "", "nREPL connection details in form of host:port. Required.")
	flag.Parse()

	if nrepl == "" {
		fmt.Println("nrepl is a required parameter")
		return
	}

	conn, err := net.Dial("tcp", nrepl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// sigchan := make(chan os.Signal, 2)
	// signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	// go func() {
	// 	<-sigchan
	// 	conn.Close()
	// 	os.Exit(1)
	// }()

	eastwood(namespaces, conn)

	if file == "-" {
		source_file_bytes, _ := ioutil.ReadAll(os.Stdin)
		b, _ := edn.Marshal(string(source_file_bytes))
		kibit(string(b), false, conn)
	} else {
		kibit(file, true, conn)
	}
}
