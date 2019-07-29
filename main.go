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
	"github.com/fatih/structs"
	"github.com/influxdata/influxdb/client/v2"
	influxClient "github.com/influxdata/influxdb/client/v2"
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
	// create db
	dbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: os.Getenv("DB_URL"),
	})
	if err != nil {
		log.Fatalf("can not connect to influxdb %v", err)
	}
	defer dbClient.Close()
	// create workers
	var workers []Worker
	job := make(chan string, maxWorker)
	quit := make(chan int)
	forever := make(chan int)
	jobResult := make(chan Signal)

	for i := 0; i < maxWorker; i++ {
		workers = append(workers, NewProductWorker(strconv.Itoa(i)))
	}

	go func() {
		for _, worker := range workers {
			res := worker.RunJob(job, quit)
			go func() {
				for {
					jobResult <- <-res
				}
			}()
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

	var client = &http.Client{
		Timeout: time.Second * 10,
	}

	go func() {
		for _, category := range categories {
			go func(c Category) {
				link := fmt.Sprintf(ProductCategoryLink, c.ID, 1)
				res, err := client.Get(link)
				if err != nil {
					log.Errorf("link %s error: %v", link, err)
					return
				}
				defer res.Body.Close()
				buf, err := ioutil.ReadAll(res.Body)
				pagingBuf, _, _, err := jsonparser.Get(buf, "result", "meta_data")
				if err != nil {
					log.Errorf("can not get paging info of %s error: %v", link, err)
					return
				}
				var paging Page
				err = json.Unmarshal(pagingBuf, &paging)
				if err != nil {
					log.Errorf("can not parser %v", err)
					return
				}
				for p := 1; p <= paging.TotalPage; p++ {
					linkPage := fmt.Sprintf(ProductCategoryLink, c.ID, p)
					log.Infof("link: %s - page: %d - total: %d", linkPage, p, paging.TotalPage)
					job <- linkPage
				}
			}(category)
		}
	}()

	// Create a new point batch
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database: "sendo_price",
	})

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			result := <-jobResult
			if result.Err != nil {
				log.Errorf("link %s - error: %v", result.Link, result.Err)
				continue
			}
			var products []Product
			err = json.Unmarshal(result.Result, &products)
			if err != nil {
				log.Errorf("can not parse products %v", err)
				continue
			}
			for _, product := range products {
				tags := map[string]string{
					"ProductID": strconv.FormatInt(product.ID, 10),
				}
				pt, err := influxClient.NewPoint("products", tags, structs.Map(product), time.Now())
				if err != nil {
					log.Error(err)
					continue
				}
				bp.AddPoint(pt)
			}
			if err := dbClient.Write(bp); err != nil {
				log.Error(err)
				continue
			}
			log.Infof("worker: %s link: %s - done", result.WorkerName, result.Link)
		}
	}()
	<-forever
}
