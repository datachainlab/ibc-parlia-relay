# TODO should be shared?
BSC_IBC_PROTO ?= ../lcp-parlia/proto/definitions
THIRD_PARTY_PROTO ?= ../ethereum-lcp-poc/go

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	@rm -rf ./proto/ibc && cp -a $(BSC_IBC_PROTO)/ibc ./proto/ && cp -a $(THIRD_PARTY_PROTO)/third_party ./
	docker run -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen:v0.3 sh ./scripts/protocgen.sh
