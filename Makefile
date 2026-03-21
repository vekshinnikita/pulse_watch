.SILENT:

ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

run: 
	go run $(ROOT_DIR)cmd/pulse_watch/main.go

create-migration: 
	$(ROOT_DIR)scripts/create_migration.sh $(filter-out $@,$(MAKECMDGOALS))

migrate:
	$(ROOT_DIR)scripts/migrate.sh $(filter-out $@,$(MAKECMDGOALS))


%:
	@: