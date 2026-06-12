.PHONY: up down run download regenerate-seed import import-sample tidy test

DB_URL ?= postgres://geo:geo@localhost:5432/geo?sslmode=disable
DATA   ?= data

up:               ## start PostGIS
	docker compose up -d db

down:
	docker compose down

run:              ## run the API server (auto-migrates on start)
	GEO_DATABASE_URL="$(DB_URL)" go run ./cmd/server

download:         ## download GADM levels 1-3 GeoJSON into ./data
	./scripts/download_gadm.sh $(DATA)

regenerate-seed:  ## re-download GSO data and regenerate seed SQL
	./scripts/regenerate_seed.sh

# Import the real GADM dataset (run `make download` first).
import:
	GEO_DATABASE_URL="$(DB_URL)" go run ./cmd/importgeo \
		-l1 $(DATA)/gadm41_VNM_1.json \
		-l2 $(DATA)/gadm41_VNM_2.json \
		-l3 $(DATA)/gadm41_VNM_3.json

# Import the tiny bundled sample (HCMC slice) for a quick smoke test.
import-sample:
	GEO_DATABASE_URL="$(DB_URL)" go run ./cmd/importgeo \
		-l1 scripts/sample_VNM_1.json \
		-l2 scripts/sample_VNM_2.json \
		-l3 scripts/sample_VNM_3.json

tidy:
	go mod tidy

test:
	go test ./...
