.PHONY:build
build:
	make -C bsc build

.PHONY:network
network:
	make -C bsc bootstrap
	make -C bsc network

.PHONY:network-down
network-down:
	make -C bsc network-down
