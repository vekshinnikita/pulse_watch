#!/bin/bash

SCRIPT_DIR=$(dirname "$0")
CONFIG_FILE="$SCRIPT_DIR/../configs/config.yaml"
ENV_FILE="$SCRIPT_DIR/../.env"
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

set -a
source $ENV_FILE
set +a

# читаем порт из YAML
DB_PORT=$(yq e '.database.port' $CONFIG_FILE)
DB_HOST=$(yq e '.database.host' $CONFIG_FILE)
DB_NAME=$(yq e '.database.db_name' $CONFIG_FILE)
DB_USER=$(yq e '.database.username' $CONFIG_FILE)
DB_PASSWORD=$DB_PASSWORD

CONNECTION_STRING="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"


echo "Применение миграций"
migrate -path $MIGRATIONS_DIR -database $CONNECTION_STRING $@