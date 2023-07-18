.PHONY:deploy
deploy:
	npm i
	npx truffle migrate --network bsc_local2 --reset
	npx truffle migrate --network bsc_local --reset

.PHONY:test
test:
	npx truffle exec apps/send.js --network bsc_local2
	npx truffle exec apps/recv.js --network bsc_local
	npx truffle exec apps/ack.js --network bsc_local2
