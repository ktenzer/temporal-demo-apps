package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var logger *log.Logger
var once sync.Once

func GetLoggerInstance() *log.Logger {
	once.Do(func() {
		logger = createLogger()
	})
	return logger
}

func createLogger() *log.Logger {
	mw := io.MultiWriter(os.Stdout)

	return log.New(mw, "", 0)
}

func LogApi(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}
