.PHONY:build
build:
	docker compose -f docker-compose.bsc.yml build
	docker compose -f docker-compose.simple.yml build --progress=plain

.PHONY:bootstrap
bootstrap:
	docker compose -f docker-compose.simple.yml run --rm bootstrap-simple

.PHONY:network
network:
	docker compose -f docker-compose.simple.yml up -d autoheal
	docker compose -f docker-compose.simple.yml up -d bsc-rpc bsc-validator1-1

.PHONY:network-down
network-down:
	docker compose -f docker-compose.simple.yml down -v
