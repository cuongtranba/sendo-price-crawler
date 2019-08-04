package main

type category struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type product struct {
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

type page struct {
	TotalPage int `json:"total_page"`
}
