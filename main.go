package main

import (
	"fmt"
	"os"
	"log"
	"time"
	"context"
	"math/big"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
)


func main() {
	
	e, err := ethclient.Dial("/home/m/.ethereum/geth.ipc")

	if err != nil {
		fmt.Printf("err %s\n", err)
	}

	cbn, err := e.BlockNumber(context.Background())
	fmt.Printf("current block: %d\n", cbn)

	// from current, back to the beginning
	for bn := cbn; bn != 0; bn-- {
		// fetch the block
		block, err := e.BlockByNumber(context.Background(), big.NewInt(int64(bn)))
		if err == nil {
			stamp := time.Unix(int64(block.Time()), 0)
			fmt.Printf("block: %d time: %s\n", bn, stamp)

			// look at each transaction in the block
			trs := block.Transactions()
			for i, t := range trs {
				// check if it was sent to the address we are interested in.
				if t.To() != nil && t.To().Hex() == "0x9F12b0E66c3E44C30e70530217B7682F5C67BA51" {
					fmt.Printf("transaction %d -- to: %s data: %s\n", i, t.To().Hex(), t.Data())
				}
			}
		}
	}

	//
	// lets send a transaction with an IPFS CID as a payload.
	//
	var cid []byte = []byte(os.Args[1])
	//
	// we send here for testing purposes, others would know to look at this for finding CIDs
	//
	toAddress := common.HexToAddress("0x9F12b0E66c3E44C30e70530217B7682F5C67BA51")


	chainID, _ := e.NetworkID(context.Background())

	// fetch our private key from an env variable
	privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATEKEY"))
	if err != nil {
		log.Fatal(err)
	}
	
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	value := big.NewInt(0)
	gasLimit := uint64(31000)
	gasPrice, err := e.SuggestGasPrice(context.Background())
	nonce, err := e.PendingNonceAt(context.Background(), fromAddress)

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, cid)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = e.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Printf("error sending: %s %v\n", err, signedTx)
	} else {
		fmt.Printf("sent: %v\n", signedTx)
	}
}
