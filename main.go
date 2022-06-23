package main

import (
	"block_demo/block"
	"fmt"
)

func main() {

	//创世纪块 初始化区块链
	bc := block.NewBlockchain()

	bc.GetBalance("lucbine") //获得矿工的余额

	bc.Send("lucbine", "xiaoyunduo", 1, "拿去花吧，随便花")

	bc.GetBalance("lucbine")
	bc.GetBalance("xiaoyunduo")

	bc.GetBalance(block.Miner)

	//bc.AddBlock([]*block.Transaction{
	//
	//})
	//
	//bc.AddBlock([]*block.Transaction{
	//
	//})

	//遍历

	bci := bc.Iterator()

	for {
		b, next := bci.PreBlock()
		fmt.Printf("prev. hash:%x\n", b.PrevBlockHash)
		fmt.Printf("hash:%x\n", b.Hash)
		fmt.Printf("time:%d\n", b.Timestamp)
		fmt.Printf("nonce:%d\n", b.Nonce)

		fmt.Printf("data:%s\n", b.Transactions[0].Vin[0].FromAddr)

		fmt.Printf(":%s\n", b.Transactions[0].Vin[0].FromAddr)

		pow := block.NewProofOfWork(b)
		fmt.Printf("pow: %t\n", pow.Validate())
		fmt.Println()

		if !next { //代表已经是创世纪块了
			break
		}

	}

}
