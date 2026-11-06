package controller

import (
	"context"
	"time"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
)

type Controller struct {
	config configuration.Config
	mongo  *mongo.Client
	pg     *pgxpool.Pool

	gatewayHealthStaleAge time.Duration
}

func New(config configuration.Config) (ctrl *Controller, err error) {
	if err = validateConfig(config); err != nil {
		return nil, err
	}
	staleAge, err := gatewayHealthStaleAge(config)
	if err != nil {
		return nil, err
	}
	mongoClient, err := connect(context.Background(), clientOptions(config), config.MongoDatabase, 10*time.Second)
	if err != nil {
		return nil, err
	}

	pool, err := newPostgresPool(context.Background(), config)
	if err != nil {
		mongoClient.Disconnect(context.Background())
		return nil, err
	}

	return &Controller{
		mongo:  mongoClient,
		pg:     pool,
		config: config,

		gatewayHealthStaleAge: staleAge,
	}, nil
}

func (this *Controller) Close() {
	this.config.GetLogger().Info("close mongo connection", "result", this.mongo.Disconnect(nil))
	this.pg.Close()
	this.config.GetLogger().Info("closed postgres pool")
}

func (this *Controller) Config() *configuration.Config {
	return &this.config
}
