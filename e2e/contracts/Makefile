.PHONY:deploy
deploy:
	npm i
	npx hardhat run ./scripts/deploy.js --network bsc_local1
	npx hardhat run ./scripts/deploy.js --network bsc_local2

.PHONY:test
test:
	npx hardhat run ./scripts/send.js --network bsc_local2
	sleep 60
	../testrly tx relay ibc01 --home ../.testrly
	sleep 20
	npx hardhat run ./scripts/recv.js --network bsc_local1
	sleep 20
	../testrly query unrelayed-packets ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
	../testrly tx relay-acknowledgements ibc01 --home ../.testrly
	sleep 20
	../testrly query unrelayed-acknowledgements ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
