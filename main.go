package main

import (
	"github.com/google/uuid"
	bencode "github.com/jackpal/bencode-go"
	"gopkg.in/edn.v1"
	"regexp"
	"fmt"
	"net"
	"os"
	"io/ioutil"
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
	Ex string
	Value string
	Status []string
	Id string
	Session string
	Out string
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

	if result.Out != ""  {
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

func eastwood(args []string, conn net.Conn){
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
	"op": "eval",
	"code": code,
	"id": msgid,
    }
    err = bencode.Marshal(conn, instruction)

    if err != nil {
	fmt.Println(err)
	return
    }

    printmsgid(conn, msgid)
}

func kibit(input string, conn net.Conn) {
    code := fmt.Sprintf(`(do (require 'kibit.check) (run! (fn [{:keys [file line expr alt]}] (printf "%%s:%%s: Consider using: %%s Instead of %%s" file line alt expr)) (kibit.check/check-reader (java.io.StringReader. %v))))`, input)

    msguuid, _ := uuid.NewRandom()
    msgid := msguuid.String()

    instruction := map[string]interface{}{
	"op": "eval",
	"code": code,
	"id": msgid,
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

	if result.Out != ""  {
	    fmt.Print(result.Out)
	}

	if len(result.Status) > 0 {
	    return
	}
    }
}

func main() {
    args := os.Args
    if len(args) < 2 {
	return
    }
    nrepl := args[1]

    conn, err := net.Dial("tcp", nrepl)
    if err != nil {
	fmt.Println(err)
	return
    }
    defer conn.Close()

    eastwood(args[2:], conn)

    source_file_bytes, err := ioutil.ReadAll(os.Stdin)
    b, err := edn.Marshal(string(source_file_bytes))
    kibit(string(b), conn)
}
