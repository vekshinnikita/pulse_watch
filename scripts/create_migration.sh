#!/bin/bash

SCRIPT_DIR=$(dirname "$0")
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

MIGRATION_NAME=$1

echo "Создание пустой миграции"
migrate create -ext sql -dir $MIGRATIONS_DIR -seq $MIGRATION_NAME