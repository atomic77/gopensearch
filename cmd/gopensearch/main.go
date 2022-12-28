package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/server"
)

func main() {

	dbLoc := flag.String("db", "test.db", "Location of sqlite database")
	port := flag.Int("port", 8080, "Port to listen on")
	listenAddr := flag.String("listenAddr", "0.0.0.0", "Address to listen on")
	debug := flag.Bool("debug", false, "Whether to produce more debugging output")
	flag.Parse()

	s := &server.Server{
		Cfg: server.Config{
			DbLocation: *dbLoc,
			ListenAddr: *listenAddr,
			Port:       *port,
			Debug:      *debug,
		},
	}
	s.Init()
	addr := fmt.Sprintf("%s:%d", s.Cfg.ListenAddr, s.Cfg.Port)
	log.Println(repr.String(s.Cfg))
	log.Println("Starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, s.Router))
}
