const fs = require('fs');
const Wallet = require('ethereumjs-wallet').default;

const args = process.argv.slice(2);
if (args.length < 1) {
  console.error("specify the hex string of the private key");
  process.exit(1);
}

const key = Buffer.from(args[0], 'hex');
const wallet = Wallet.fromPrivateKey(key);
wallet.toV3String('')
  .then(s => {
    const addr = wallet.getAddressString()
    fs.writeFile(`./${addr}`, s, function(err){
      if (err) {
        throw err;
      }
    })
  }).catch(e => {
    console.error(e);
    process.exit(1);
});
