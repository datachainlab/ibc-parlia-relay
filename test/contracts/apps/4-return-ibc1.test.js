const MiniToken = artifacts.require("MiniToken");

contract("MiniToken", (accounts) => {
  it("should put 0 MiniToken in bob account on ibc1", () =>
    MiniToken.deployed()
      .then((instance) => instance.balanceOf(accounts[2]))
      .then((balance) => {
          console.log("amount = ", balance.toString())
        assert.equal(balance.valueOf(), 0, "50 wasn't in Bob account");
      }));
});
