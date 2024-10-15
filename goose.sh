#!/bin/bash

source .env

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASS dbname=$DB_NAME sslmode=$DB_SSL_MODE TimeZone=$DB_TIME_ZONE"

goose -dir app/external/sqlc/migrations $@