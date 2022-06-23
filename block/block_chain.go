package block

import (
	"encoding/hex"
	"fmt"

	"github.com/boltdb/bolt"
)

const (
	DBFile              = "blockchain.db"
	BlocksBucket        = "blocks-1"
	LastBlockKey        = "last"
	Miner               = "bubu"
	GenesisCoinbaseData = "The times 2022/06/22 chancellor on brink of second bailout for banks"
)

type Blockchain struct {
	lastBlockHash []byte //用来记录区块的hash 值
	db            *bolt.DB
}

func NewBlockchain() *Blockchain {
	var lastBlockHash []byte

	//打开数据库
	db, _ := bolt.Open(DBFile, 0600, nil)
	//跟新数据

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlocksBucket))
		if bucket == nil {
			//第一次使用，创建创世纪块
			fmt.Println("No existing blockchain found . Creating a new one...")

			genesis := NewGenesisBlock(NewCoinbaseTX(Miner, GenesisCoinbaseData))
			blockData := genesis.Serialize()
			bucket, _ = tx.CreateBucket([]byte(BlocksBucket))
			bucket.Put(genesis.Hash, blockData)
			lastBlockHash = genesis.Hash
			bucket.Put([]byte(LastBlockKey), lastBlockHash)
		} else {
			lastBlockHash = bucket.Get([]byte(LastBlockKey))
		}
		return nil
	})

	return &Blockchain{
		lastBlockHash: lastBlockHash,
		db:            db,
	}
}

//添加区块

func (bc *Blockchain) MineBlock(txs []*Transaction, data string) {
	var lastBlockHash []byte

	//获得最新区块
	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlocksBucket))
		lastBlockHash = bucket.Get([]byte(LastBlockKey))
		return nil
	})

	//获得最新的区块hash

	bc.db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(BlocksBucket))

		//创建coinbase 交易
		cbtx := NewCoinbaseTX(Miner, data)

		txs = append(txs, cbtx) //打包所有的交易

		block := NewBlock(txs, lastBlockHash)

		//将新区块放入到db
		bucket.Put(block.Hash, block.Serialize())

		bucket.Put([]byte(LastBlockKey), block.Hash)

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
		bucket := tx.Bucket([]byte(BlocksBucket))

		encodeBlock := bucket.Get(bci.currentHash)

		block = UnSerialize(encodeBlock)

		return nil
	})

	//当前hash变更为当前hash
	bci.currentHash = block.PrevBlockHash

	return block, len(bci.currentHash) > 0

}

// 查找可以解锁的全部交易

func (bc *Blockchain) FindUnSpentTransactions(address string) []*Transaction {
	var unspentTXs []*Transaction //未花费的交易输出

	//已经花出的utxo 构建tx -> voutIdx 的map

	spentTXOS := make(map[string][]int)

	bci := bc.Iterator()

	for {
		block, next := bci.PreBlock()

		//fmt.Println("FindUnSpentTransactions block hash: ", hex.EncodeToString(block.Hash))

		for _, tx := range block.Transactions {

			//https://blog.csdn.net/dengpost/article/details/109703427
			txId := hex.EncodeToString(tx.ID) //找到交易hash

			//fmt.Printf("FindUnSpentTransactions block hash: %s,transaction hash %s ;isCoinbase : %t \n", hex.EncodeToString(block.Hash), txId, tx.IsCoinbase())

		Output:
			for outIdx, out := range tx.Vout {
				//如果已经被花费 ，直接跳过该交易
				if spentTXOS[txId] != nil {
					for _, spentOut := range spentTXOS[txId] { //交易输出项目
						if spentOut == outIdx {
							continue Output
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			//用来维护spentTXOS ,已经被引用过了 ，代表被使用

			if !tx.IsCoinbase() { //非 coinbase 交易
				for _, in := range tx.Vin { //交易输入项
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxId)
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.VoutIdx)
					}
				}
			}
		}

		if !next { //如果是genesis 遍历结束
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOS []TXOutput

	unspentTransactions := bc.FindUnSpentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOS = append(UTXOS, out)
			}
		}
	}
	return UTXOS
}

func (bc *Blockchain) GetBalance(address string) {
	balance := 0

	UTXOS := bc.FindUTXO(address)

	for _, out := range UTXOS {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s':%d\n", address, balance)

}

//获得满足交易的UTXO

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {

	var (
		unspendOutputs = make(map[string][]int)
		accumulated    = 0
		unspendTXs     = bc.FindUnSpentTransactions(address)
	)

work:
	for _, tx := range unspendTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspendOutputs[txID] = append(unspendOutputs[txID], outIdx)

				//utxo 足够用跳出循环

				if accumulated >= amount {
					break work
				}
			}
		}
	}
	return accumulated, unspendOutputs
}

func (bc *Blockchain) Send(from, to string, amount int, data string) {
	tx := NewUTXOTransaction(from, to, amount, bc)

	bc.MineBlock([]*Transaction{tx}, data)

	fmt.Println("blockchain transaction success!")

}

//func dbExists() bool {
//	if _, err := os.Stat(DBFile); os.IsNotExist(err) {
//		return false
//	}
//	return true
//}
