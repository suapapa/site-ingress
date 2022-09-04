package main

import (
	"log"
	"net/http"
	"os"
)

type AcmeChallenge struct {
	fileHandler http.Handler
}

func NewAcmeChallenge(acPath string) *AcmeChallenge {
	os.MkdirAll(acPath, 0700)
	return &AcmeChallenge{
		fileHandler: http.FileServer(http.FileSystem(http.Dir(acPath))),
	}
}

func (ac *AcmeChallenge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ac.fileHandler.ServeHTTP(w, r)

	// TODO: It try to start before certfile created
	log.Println("(re)start https server")
	startHTTPSServer()
}
