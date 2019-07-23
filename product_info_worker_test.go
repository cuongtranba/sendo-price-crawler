package main

import "testing"

func TestProductInfoWorker_RunJob(t *testing.T) {
	type fields struct {
		name string
	}
	type args struct {
		job    <-chan string
		quit   <-chan int
		signal <-chan Signal
	}

	job := make(chan string)
	job <- "https://www.sendo.vn/m/wap_v2/category/product?category_id=2075&listing_algo=algo5&p=1&platform=wap&s=30&sortType=default_listing_desc"
	close(job)
	quit := make(chan int)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "get product done",
			fields: fields{
				name: "product worker",
			},
			args: args{
				job:  job,
				quit: quit,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw := &ProductInfoWorker{
				name: tt.fields.name,
			}
			pw.RunJob(tt.args.job, tt.args.quit, tt.args.signal)
		})
	}
}
