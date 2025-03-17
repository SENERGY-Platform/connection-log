package controller

import (
	"context"
	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	influx "github.com/influxdata/influxdb1-client/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Controller struct {
	config    configuration.Config
	mongo     *mongo.Client
	influx    influx.Client
	queries   *queryTemplates
	influxUTC bool
}

func New(config configuration.Config) (ctrl *Controller, err error) {
	qt, err := newQueryTemplates()
	if err != nil {
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		return nil, err
	}
	err = createIndexes(config, mongoClient)
	if err != nil {
		mongoClient.Disconnect(context.Background())
		return nil, err
	}

	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     config.InfluxdbUrl,
		Username: config.InfluxdbUser,
		Password: config.InfluxdbPw,
		Timeout:  time.Duration(config.InfluxdbTimeout) * time.Second,
	})
	if err != nil {
		mongoClient.Disconnect(context.Background())
		return nil, err
	}

	return &Controller{
		mongo:     mongoClient,
		influx:    influxClient,
		config:    config,
		queries:   qt,
		influxUTC: config.InfluxdbUseUTC,
	}, nil
}

func (this *Controller) Close() {
	log.Println("close mongo connection:", this.mongo.Disconnect(nil))
	log.Println("close influx connection:", this.influx.Close())
}
