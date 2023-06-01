.PHONY:build
build:
	docker compose -f docker-compose.bsc.yml build
	docker compose -f docker-compose.simple.yml build --progress=plain --no-cache

.PHONY:bootstrap
bootstrap:
	docker compose -f docker-compose.simple.yml run --rm bootstrap-simple
	docker compose -f docker-compose.simple.yml run --rm bootstrap-simple2

.PHONY:network
network:
	docker compose -f docker-compose.simple.yml up -d bsc-rpc bsc-validator1-1 bsc-validator1-2 bsc-validator1-3 bsc-validator1-4 bsc-validator1-5
	docker compose -f docker-compose.simple.yml up -d bsc-rpc2 bsc-validator2-1 bsc-validator2-2 bsc-validator2-3 bsc-validator2-4 bsc-validator2-5

.PHONY:network-down
network-down:
	docker compose -f docker-compose.simple.yml down -v
