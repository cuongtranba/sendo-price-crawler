package main

import (
	"os"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	database = "nodes"
	username = "monitor"
	password = "secret"
)

// InfluxDb InfluxDb database
type InfluxDb struct {
	client client.Client
}

// NewInfluxDbClient NewInfluxDbClient
func NewInfluxDbClient(con string) (*InfluxDb, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: os.Getenv("DB_URL"),
	})
	if err != nil {
		return nil, err
	}
	return &InfluxDb{
		client: c,
	}, nil
}

// Insert insert product measurement
func (c *InfluxDb) Insert(products []Product) {

}
