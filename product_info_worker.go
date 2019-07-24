package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	jsonparser "github.com/buger/jsonparser"
	log "github.com/sirupsen/logrus"
)

type Category struct {
	ID      string `json:"id"`
	URLPath string `json:"url_path"`
}

type Product struct {
	ID               int64  `json:"id"`
	ProductID        int64  `json:"product_id"`
	Name             string `json:"name"`
	Price            int64  `json:"price"`
	PriceMax         int64  `json:"price_max"`
	FinalPrice       int64  `json:"final_price"`
	FinalPriceMax    int64  `json:"final_price_max"`
	PromotionPercent int    `json:"promotion_percent"`
	IMG              string `json:"img_url"`
}

// ProductInfoWorker get product info worker
type ProductInfoWorker struct {
	name string
}

// RunJob get product info from link
func (pw *ProductInfoWorker) RunJob(job <-chan string, quit <-chan int) <-chan Signal {
	reportSignal := make(chan Signal)

	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	go func() {
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

				var products []Product
				productBuf, _, _, err := jsonparser.Get(buf, "result", "data")
				err = json.Unmarshal(productBuf, &products)
				if err != nil {
					reportSignal <- Signal{
						Sig: ErrorSig,
						Err: err,
					}
					continue
				}
				reportSignal <- Signal{
					Err:    nil,
					Sig:    DoneSig,
					Result: products,
				}
			case <-quit:
				log.Infof("worker %s stop", pw.name)
				return
			}
		}
	}()

	return reportSignal
}

// NewProductWorker create product worker
func NewProductWorker(name string) Worker {
	return &ProductInfoWorker{
		name: name,
	}
}
