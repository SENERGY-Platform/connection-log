Provides HTTP-API to request current and historical connection log. 
Historical data is read from a influxdb instance.
Current data is read form a mongodb instance.
The data is written by the connectionlog-worker service.

Generate swagger docs:

    swag init -g api.go -o docs -dir pkg/api --parseDependency --ot json