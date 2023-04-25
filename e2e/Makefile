.PHONY:chain
chain:
	make -C chains build
	make -C chains network

.PHONY:contracts
contracts:
	make -C contracts deploy

.PHONY:relayer
relayer:
	rm -rf .testrly
	# go build -tags dev -ldflags="-X github.com/datachainlab/ibc-parlia-relay/module/constant.epochBlockPeriod=20" -o testrly .
	go build -o testrly .
	./testrly config init --home .testrly
	./testrly chains add-dir config/demo/ --home .testrly
	./testrly paths add ibc0 ibc1 ibc01 --file=config/path.json --home .testrly
	# Must wait for sufficient blocks (max 400 blocks) after contract creation
	echo "Waiting contracts finalized"
	sleep 1200
	./testrly tx clients ibc01 --home .testrly
	./testrly tx update-clients ibc01 --home .testrly
	./testrly tx connection ibc01 --home .testrly
	./testrly tx channel ibc01 --home .testrly
	nohup ./testrly service start ibc01 --home .testrly > testrly.log  &

.PHONY:test
test:
	make -C contracts test

.PHONY:down
down:
	make -C chains network-down
