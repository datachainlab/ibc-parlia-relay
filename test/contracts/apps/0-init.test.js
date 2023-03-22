const SimpleToken = artifacts.require("@hyperledger-labs/yui-ibc-solidity/SimpleToken");

contract("SimpleToken", (accounts) => {
  it("should put 100 MiniToken in Alice account on ibc0", () =>
      SimpleToken.deployed()
      .then((instance) => instance.balanceOf(accounts[1]))
      .then((balance) => {
          console.log("token amount = ", balance.toString())
          assert.equal(balance.valueOf(), 100, "100 wasn't in Alice account");
      }));
});
