package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type category struct {
	ID      string `json:"id"`
	URLPath string `json:"url_path"`
}

type product struct {
	ID               string `json:"id"`
	ProductID        string `json:"product_id"`
	UID              string `json:"uid"`
	Name             string `json:"name"`
	Price            int64  `json:"price"`
	PriceMax         int64  `json:"price_max"`
	FinalPrice       int64  `json:"final_price"`
	FinalPriceMax    int64  `json:"final_price_max"`
	PromotionPercent int    `json:"promotion_percent"`
}

// ProductInfoWorker get product info worker
type ProductInfoWorker struct {
	name string
}

// RunJob get product info from link
func (pw *ProductInfoWorker) RunJob(job <-chan string, quit <-chan int, reportSignal chan<- Signal) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	for {
		select {
		case link := <-job:
			response, err := netClient.Get(link)
			if err != nil {
				reportSignal <- Signal{
					Sig: ErrorSig,
					Err: err,
				}
				continue
			}

			buf, err := ioutil.ReadAll(response.Body)
			if err != nil {
				reportSignal <- Signal{
					Sig: ErrorSig,
					Err: err,
				}
				continue
			}

			var products []product
			err = json.Unmarshal(buf, &products)
			if err != nil {
				reportSignal <- Signal{
					Sig: ErrorSig,
					Err: err,
				}
				continue
			}

			log.Info(products)
		case <-quit:
			log.Infof("worker %s stop", pw.name)
			return
		}
	}
}

// NewProductWorker create product worker
func NewProductWorker(name string) Worker {
	return &ProductInfoWorker{
		name: name,
	}
}
