package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic1", startPanic1)
	mux.HandleFunc("/panic", startPanic)
	mux.HandleFunc("/", hello)
	fmt.Println("Server listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", recoverFromPanic(mux)))
}

func recoverFromPanic(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			re := recover()
			if re != nil {
				log.Println(re)
				stack := debug.Stack()
				//log.Println(string(stack))
				path, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(path)
				re := regexp.MustCompile(`\/[]\/[a-z0-9A-Z]+\.go:[0-9]+`)
				s := re.FindAll([]byte(string(stack)), -1)
				for _, bs := range s {
					filePath := strings.Split(string(bs), ":")[0]
					line := strings.Split(string(bs), ":")[1]
					lineNumber, err := strconv.Atoi(line)
					if err != nil {
						lineNumber = -1
					}
					var lines [][2]int
					if lineNumber > 0 {
						lines = append(lines, [2]int{lineNumber, lineNumber})
					}
					b, err := ioutil.ReadFile(filePath)
					if err != nil {
						fmt.Println(err)
					}
					lexer := lexers.Get("go")
					iterator, err := lexer.Tokenise(nil, string(b))
					style := styles.Get("github")
					if style == nil {
						style = styles.Fallback
					}
					formatter := html.New(html.TabWidth(2), html.WithLineNumbers(true), html.LineNumbersInTable(true), html.HighlightLines(lines))
					w.Header().Set("Content-type", "text/html")
					fmt.Fprintf(w, "<style> pre {font-size: 1.2em}</style>")
					formatter.Format(w, style, iterator)
				}
				fmt.Fprintln(w, string(stack))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		rw := &responseWriter{ResponseWriter: w}
		app.ServeHTTP(rw, r)
		rw.Flush()
	}
}

type responseWriter struct {
	http.ResponseWriter
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("Hijacking not possible")
	}
	con, bf, err := hj.Hijack()
	if err != nil {
		return nil, nil, err
	}
	return con, bf, err
}

func (r *responseWriter) Flush() {
	flusher, isErr := r.ResponseWriter.(http.Flusher)
	if !isErr {
		return
	}
	flusher.Flush()
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Namastey Duniyaa!")
}

func startPanic1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Something")
	panic("Lutia dub gayi1")
}

func startPanic(w http.ResponseWriter, r *http.Request) {
	panic("Lutia dub gayi")
}
