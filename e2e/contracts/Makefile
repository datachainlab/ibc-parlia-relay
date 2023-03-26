.PHONY:deploy
deploy:
	npm i
	npx truffle migrate --network bsc_local2 --reset
	npx truffle migrate --network bsc_local --reset

.PHONY:test
test:
	npx truffle exec apps/send.js --network bsc_local2
