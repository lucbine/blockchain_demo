package main

import (
	"block_demo/block"
	"fmt"
)

func main() {

	//创世纪块 初始化区块链
	bc := block.NewBlockchain()

	//bc.AddBlock("send 1 BTC to lucine")
	//
	//bc.AddBlock("send 2 BTC to bubu")

	//遍历

	bci := bc.Iterator()

	for {
		b, next := bci.PreBlock()
		fmt.Printf("prev. hash:%x\n", b.PrevBlockHash)
		fmt.Printf("data:%s\n", b.Data)
		fmt.Printf("hash:%x\n", b.Hash)
		fmt.Printf("time:%d\n", b.Timestamp)
		fmt.Printf("nonce:%d\n", b.Nonce)

		pow := block.NewProofOfWork(b)
		fmt.Printf("pow: %t\n", pow.Validate())
		fmt.Println()

		if !next { //代表已经是创世纪块了
			break
		}

	}

}
