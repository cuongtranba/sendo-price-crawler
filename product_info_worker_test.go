package main

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestProductInfoWorker_RunJobFail(t *testing.T) {
	type fields struct {
		name string
	}
	type args struct {
		job  <-chan string
		quit <-chan int
	}

	job := make(chan string)
	quit := make(chan int)
	returnSignal := make(chan Signal)

	go func() {
		job <- "test"
		returnSignal <- Signal{
			Sig: ErrorSig,
		}
	}()

	tests := []struct {
		name   string
		fields fields
		args   args
		want   <-chan Signal
	}{
		{
			name:   "error when wrong url",
			fields: fields{name: "get product"},
			args: args{
				job:  job,
				quit: quit,
			},
			want: returnSignal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw := &ProductInfoWorker{
				name: tt.fields.name,
			}
			got := pw.RunJob(tt.args.job, tt.args.quit)
			result := <-got
			if result.Sig != ErrorSig {
				t.Errorf("fail because expected: %v - actual %v", ErrorSig, result.Sig)
			}
		})
	}
}

func TestProductInfoWorker_RunJobDone(t *testing.T) {
	type fields struct {
		name string
	}
	type args struct {
		job  <-chan string
		quit <-chan int
	}

	job := make(chan string)
	quit := make(chan int)
	returnSignal := make(chan Signal)

	go func() {
		job <- "https://www.sendo.vn/m/wap_v2/category/product?category_id=1&listing_algo=algo5&p=%7B%7B.Page%7D%7D&platform=wap&s=30&sortType=default_listing_desc"
		returnSignal <- Signal{
			Sig: ErrorSig,
		}
	}()
	pw := &ProductInfoWorker{
		name: "worker get product",
	}
	gotCh := pw.RunJob(job, quit)
	result := <-gotCh
	if result.Err != nil {
		t.Errorf("error when process api %v", result.Err)
		return
	}
	res := result.Result.([]Product)
	log.Info(res)
}
