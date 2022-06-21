package block

import (
	"fmt"

	"github.com/boltdb/bolt"
)

const (
	dbFile = "blockchain.db"
	//bucket 名称
	blocksBucket = "blocks"
	lastBlockKey = "last"
)

type Blockchain struct {
	lastBlockHash []byte //用来记录区块的hash 值
	db            *bolt.DB
}

func NewBlockchain() *Blockchain {

	var lastBlockHash []byte

	//打开数据库
	db, _ := bolt.Open(dbFile, 0600, nil)
	//跟新数据

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			//第一次使用，创建创世纪块
			fmt.Println("No existing blockchain found . Creating a new one...")
			genesis := NewGenesisBlock()
			blockData := genesis.Serialize()
			bucket, _ = tx.CreateBucket([]byte(blocksBucket))
			bucket.Put(genesis.Hash, blockData)
			lastBlockHash = genesis.Hash
			bucket.Put([]byte(lastBlockKey), lastBlockHash)
		} else {
			lastBlockHash = bucket.Get([]byte(lastBlockKey))
		}
		return nil
	})

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		db:            db,
	}
}

//添加区块

func (bc *Blockchain) AddBlock(data string) {
	var lastBlockHash []byte

	//获得最新区块
	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastBlockHash = bucket.Get([]byte(lastBlockKey))
		return nil
	})

	//获得最新的区块hash

	bc.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blocksBucket))

		block := NewBlock(data, lastBlockHash)

		//将新区块放入到db
		bucket.Put(block.Hash, block.Serialize())

		bucket.Put([]byte(lastBlockKey), block.Hash)

		//覆盖 最新hash的值
		bc.lastBlockHash = block.Hash
		return nil
	})
}

//迭代器

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.lastBlockHash, bc.db}
	return bci
}

// 获得前一个区块hash ,返回当前区块数据

func (bci *BlockchainIterator) PreBlock() (*Block, bool) {
	var block *Block

	bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		encodeBlock := bucket.Get(bci.currentHash)

		block = UnSerialize(encodeBlock)

		return nil
	})

	//当前hash变更为当前hash
	bci.currentHash = block.PrevBlockHash

	return block, len(bci.currentHash) > 0

}
