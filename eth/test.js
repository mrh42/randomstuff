var instance = contract.at(deployed_contract.address);

instance.add.sendTransaction("mrh0", { from: eth.accounts[0], gas: 1000000 });

instance.add.sendTransaction("mrh1", { from: eth.accounts[0], gas: 1000000 });

instance.get.call(0);
instance.get.call(instance.len.call()-1);

instance.set.sendTransaction("foo", "bafkreig5o4tza7wt6pwyalbxuwukvzrzodda5muz5gulhizmx24ksasslm", { from: eth.accounts[0], gas: 1000000 });
instance.lookup.call("foo");
