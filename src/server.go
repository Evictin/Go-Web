package main 

import (
    "net/http"
    "fmt"
    "strings"
    "strconv"
    "encoding/base64"
    "time"
    "os"
)

type loggingHandler struct { http.Handler }
type loggingResponseWriter struct { 
	status int
	http.ResponseWriter
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == -1 {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(data)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<html><head><link rel='stylesheet' type='text/css' href='/style/style.css' /></head>")
	fmt.Fprintln(w, "<body>")
    fmt.Fprintf(w, "<h1>Hello %s!</h1>", r.URL.Path[1:])
	fmt.Fprintln(w, "</body></html>")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h2>Hello, User %s</h2>", r.URL.Path[2:])
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lw := &loggingResponseWriter{-1, w}
	h.Handler.ServeHTTP(lw, r)
	fmt.Printf("%s: %-40s [%d]\n", r.Method, r.URL, lw.status)
}

func basicAuth(r *http.Request) (username, password string) {
	authHeader := r.Header.Get("Authorization")
	if strings.Index(authHeader, "Basic ") == 0 {
		data, err := base64.StdEncoding.DecodeString(authHeader[6:len(authHeader)])
		if err == nil {
			substrings := strings.Split(string(data), ":")
			if len(substrings) == 2 {
				username, password = substrings[0], substrings[1]
			}
		}
	}	
	return
}

func shutdown(timeout time.Duration) {
	if interval := timeout/time.Second; interval > 0 { 
		countdown := time.Tick(time.Second)
		for ; interval > 0; interval-- {
			fmt.Printf("Server shutting down in %d seconds...\n", interval);
			<- countdown
		}
	}
	fmt.Println("Server shutdown at " + time.Now().String())
	os.Exit(0)
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	username, password := basicAuth(r)
	if username == "jesse" && password == "dragon" {
		seconds, _ := strconv.Atoi(r.FormValue("seconds"))
		if seconds <= 0 {
			seconds = 5
		}
		fmt.Fprintln(w, "<html><head><link rel='stylesheet' type='text/css' href='/style/style.css' /></head>")
		fmt.Fprintln(w, "<body><h1>Shutting down server...</h1></body></html>")
		go shutdown(time.Duration(seconds) * time.Second)
	} else {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"my realm\"")
		w.WriteHeader(401)
	}
}

func main() {
    http.Handle("/user/{id}/", &loggingHandler{http.HandlerFunc(userHandler)})
    http.Handle("/hello", &loggingHandler{http.HandlerFunc(helloHandler)})
	http.Handle("/shutdown", &loggingHandler{http.HandlerFunc(shutdownHandler)})
	http.Handle("/style/", http.StripPrefix("/style", http.FileServer(http.Dir("style"))))	// don't log access to stylesheets
    http.Handle("/", &loggingHandler{http.FileServer(http.Dir("public"))})
    http.ListenAndServe(":8080", nil)
}

