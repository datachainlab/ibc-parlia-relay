.PHONY:chain
chain:
	make -C chains build
	make -C chains network
	# Wait for one epoch to use BEP-126 validator
	sleep 60

.PHONY:contracts
contracts:
	make -C contracts deploy

.PHONY:relayer
relayer:
	rm -rf .testrly
	go build -o testrly .
	./testrly config init --home .testrly
	./testrly chains add-dir config/demo/ --home .testrly
	./testrly paths add ibc0 ibc1 ibc01 --file=config/path.json --home .testrly
	./testrly tx clients ibc01 --home .testrly
	./testrly tx update-clients ibc01 --home .testrly
	./testrly tx connection ibc01 --home .testrly
	./testrly tx channel ibc01 --home .testrly

.PHONY:test
test:
	make -C contracts test

.PHONY:down
down:
	make -C chains network-down
