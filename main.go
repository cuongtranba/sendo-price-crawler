package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/buger/jsonparser"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

const (
	maxWorker           = 10000
	productCategoryLink = "https://www.sendo.vn/m/wap_v2/category/product?category_id=%d&listing_algo=algo5&p=%d&platform=wap&s=30&sortType=default_listing_desc"
	categoryLink        = "https://www.sendo.vn/m/wap_v2/category/sitemap"
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

	// dbClient, err := client.NewHTTPClient(client.HTTPConfig{
	// 	Addr: os.Getenv("DB_URL"),
	// })
	// if err != nil {
	// 	log.Fatalf("can not connect to influxdb %v", err)
	// }
	// defer dbClient.Close()
	var wg sync.WaitGroup
	wg.Add(maxWorker)
	quit := make(chan bool)
	jobs := make(chan string, maxWorker)
	result := make(chan []product, maxWorker)
	categories, err := getCategories(categoryLink)
	if err != nil {
		log.Fatal(err)
	}

	for w := 1; w <= maxWorker; w++ {
		go do(jobs, quit, &wg, result)
	}
	wg.Add(1)
	go processResult(result, quit, &wg)

	for _, c := range categories {
		productLinks, err := getProductLinks(c)
		if err != nil {
			log.Errorf("can not get product id %d - err:%v ", c.ID, err)
			continue
		}
		for _, pl := range productLinks {
			jobs <- pl
		}
	}

	quit <- true
	wg.Wait()
}

func processResult(result <-chan []product, quit <-chan bool, wg *sync.WaitGroup) {
	count := 0
	for {
		select {
		case <-quit:
			wg.Done()
			return
		case products := <-result:
			count += len(products)
			log.Infof("process %d products", count)
		}
	}
}

func do(job <-chan string, quit <-chan bool, wg *sync.WaitGroup, result chan<- []product) {
	for {
		select {
		case <-quit:
			wg.Done()
			return
		case link := <-job:
			// log.Infof("get link %s", link)
			var products []product
			err := requestGet(link, &products, "result", "data")
			if err != nil {
				log.Errorf("can not get link %s err: %v", link, err)
				continue
			}
			result <- products
			// log.Infof("get link %s - done", link)
		}
	}
}

func getProductLinks(c category) ([]string, error) {
	link := fmt.Sprintf(productCategoryLink, c.ID, 1)
	var paging page
	err := requestGet(link, &paging, "result", "meta_data")
	if err != nil {
		return nil, err
	}
	var links []string
	for i := 1; i <= paging.TotalPage; i++ {
		links = append(links, fmt.Sprintf(productCategoryLink, c.ID, i))
	}
	return links, nil
}

func getCategories(link string) ([]category, error) {
	var categories []category
	err := requestGet(link, &categories, "result", "data")
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func requestGet(link string, model interface{}, path ...string) error {
	httpClient := &http.Client{
		Timeout: time.Second * 2,
	}

	res, err := httpClient.Get(link)
	if err != nil {
		return fmt.Errorf("can not get %s err: %v", link, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("can not read body request %v", err)
	}

	body, _, _, err = jsonparser.Get(body, path...)
	if err != nil {
		return fmt.Errorf("error when get path parser %v", err)
	}

	err = json.Unmarshal(body, &model)
	if err != nil {
		return fmt.Errorf("can not parser json %v", err)
	}
	return nil
}
