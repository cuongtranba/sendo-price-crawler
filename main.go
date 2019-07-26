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
	//create db
	// clientDb, err := NewInfluxDbClient(os.Getenv("DB_URL"))
	// if err != nil {
	// 	log.Fatal("can not create connection to influx db %v", err)
	// }
	//create workers
	var workers []Worker
	job := make(chan string)
	quit := make(chan int)
	forever := make(chan int)
	jobResult := make(chan Signal)

	for i := 0; i < maxWorker; i++ {
		workers = append(workers, NewProductWorker(strconv.Itoa(i)))
	}

	go func() {
		for _, worker := range workers {
			res := worker.RunJob(job, quit)
			jobResult <- <-res
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

	go func() {
		for _, category := range categories {
			link := fmt.Sprintf(ProductCategoryLink, category.ID, 1)
			job <- link
		}
	}()

	go func() {
		for {
			result := <-jobResult
			if result.Err != nil {
				log.Errorf("link %s - error: %v", result.Link, result.Err)
				continue
			}
			log.Infof("link: %s - done", result.Link)
		}
	}()
	<-forever
}
