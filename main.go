package main

import (
	"fmt"
	//"strings"
	"context"
	"math/big"
	//"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/accounts/abi"
	//"github.com/ethereum/go-ethereum/common"
)


func main() {
	
	e, err := ethclient.Dial("/home/m/.ethereum/geth.ipc")

	if err != nil {
		fmt.Printf("err %s\n", err)
	}

	cbn, err := e.BlockNumber(context.Background())
	fmt.Printf("current block: %d\n", cbn)

	for bn := cbn; bn != 0; bn-- {
		block, err := e.BlockByNumber(context.Background(), big.NewInt(int64(bn)))
		if err == nil {
			fmt.Printf("block: %d time: %d\n", bn, block.Time())
			
			trs := block.Transactions()
			for i, t := range trs {
				fmt.Printf("transaction %d -- to: %x data: %x\n", i, t.To(), t.Data())
			}
		}
	}
}
