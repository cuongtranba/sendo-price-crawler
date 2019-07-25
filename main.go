package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	jsonparser "github.com/buger/jsonparser"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	maxWorker = 100
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	//create workers
	var workers []Worker
	job := make(chan string)
	quit := make(chan int)
	forever := make(chan int)
	jobResult := make(chan chan Signal)
	for i := 0; i < maxWorker; i++ {
		workers = append(workers, NewProductWorker(strconv.Itoa(i)))
	}

	go func() {
		for _, worker := range workers {
			res := worker.RunJob(job, quit)
			jobResult <- res
		}
	}()

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	res, err := netClient.Get(CategoryLink)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var categories []Category
	categoryBuf, _, _, err := jsonparser.Get(buf, "result", "data")
	err = json.Unmarshal(categoryBuf, &categories)
	if err != nil {
		log.Fatal(err)
	}

	for _, category := range categories {
		link := fmt.Sprintf(ProductCategoryLink, category.ID, 1)
		job <- link
	}
}
