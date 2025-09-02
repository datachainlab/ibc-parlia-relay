.PHONY:deploy
deploy:
	npm i
	npx hardhat run ./scripts/deploy.js --network bsc_local1
	npx hardhat run ./scripts/deploy.js --network bsc_local2

.PHONY:test
test:
	npx hardhat run ./scripts/send.js --network bsc_local2
	sleep 10
	../testrly tx relay ibc01 --home ../.testrly
	sleep 10
	npx hardhat run ./scripts/recv.js --network bsc_local1
	sleep 10
	../testrly query unrelayed-packets ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
	../testrly tx relay-acknowledgements ibc01 --home ../.testrly
	sleep 10
	../testrly query unrelayed-acknowledgements ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
