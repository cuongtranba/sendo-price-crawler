package main

import (
	"io/ioutil"
	"net/http"
	"time"

	jsonparser "github.com/buger/jsonparser"
	log "github.com/sirupsen/logrus"
)

// Category Category
type Category struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Image string `json:"image"`
}

// Product Product
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

// Page paging
type Page struct {
	CurrentPage int `json:"current_page"`
	TotalPage   int `json:"total_page"`
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
				// log.Infof("get job: %s", link)
				response, err := netClient.Get(link)
				if err != nil {
					reportSignal <- Signal{
						Sig:        ErrorSig,
						Err:        err,
						Link:       link,
						Result:     nil,
						WorkerName: pw.name,
					}
					continue
				}
				defer response.Body.Close()
				buf, err := ioutil.ReadAll(response.Body)
				if err != nil {
					reportSignal <- Signal{
						Sig:        ErrorSig,
						Err:        err,
						Link:       link,
						Result:     nil,
						WorkerName: pw.name,
					}
					continue
				}

				value, _, _, err := jsonparser.Get(buf, "result", "data")
				if err != nil {
					reportSignal <- Signal{
						Sig:        ErrorSig,
						Err:        err,
						Link:       link,
						Result:     nil,
						WorkerName: pw.name,
					}
					continue
				}
				reportSignal <- Signal{
					Err:        nil,
					Sig:        DoneSig,
					Result:     value,
					Link:       link,
					WorkerName: pw.name,
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
