.PHONY:deploy
deploy:
	npm i
	npx truffle migrate --network bsc_local2 --reset
	npx truffle migrate --network bsc_local --reset

.PHONY:test
test:
	npx truffle exec apps/send.js --network bsc_local2
	sleep 60
	../testrly tx relay ibc01 --home ../.testrly
	npx truffle exec apps/recv.js --network bsc_local
	../testrly query unrelayed-packets ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
	../testrly tx relay-acknowledgements ibc01 --home ../.testrly
	../testrly query unrelayed-acknowledgements ibc01 --home ../.testrly | grep '{"src":\[\],"dst":\[\]}'
