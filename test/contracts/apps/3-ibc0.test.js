const MiniToken = artifacts.require("MiniToken");

contract("MiniToken", (accounts) => {
  it("should have 50 MiniToken in alice account on ibc0", () =>
    MiniToken.deployed()
      .then((instance) => instance.balanceOf(accounts[1]))
      .then((balance) => {
          console.log("amount = ", balance.toString())
        assert.equal(balance.valueOf(), 50, "50 wasn't in Alice account");
      }));
});
