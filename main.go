package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	maxWorker = 100
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	//create workers
	var workers []Worker
	for i := 0; i < maxWorker; i++ {
		workers = append(workers, NewProductWorker(strconv.Itoa(i)))
	}

	// var category Category

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

}
