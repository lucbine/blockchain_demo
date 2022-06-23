package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64          //时间
	PrevBlockHash []byte         //前块hash值
	Hash          []byte         //当前块hash值
	Nonce         int64          //随机值
	Transactions  []*Transaction //交易信息
}

//创建创世纪块

func NewGenesisBlock(coinbase *Transaction) *Block {
	//给用户奖励
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//创建区块

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  txs,
		PrevBlockHash: prevBlockHash,
	}

	//需要先进行挖坑
	pow := NewProofOfWork(block)

	//开始挖坑，计算hash 值
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce
	return block
}

//设置头部hash

func (b *Block) SetHash() {
	tm := []byte(strconv.FormatInt(b.Timestamp, 10))

	txIds := make([][]byte, 0)

	for _, tx := range b.Transactions {
		txIds = append(txIds, tx.ID)
	}

	mergeHash := append(txIds, tm, b.PrevBlockHash)

	//将前一个hash、交易信息、时间戳 合并
	headers := bytes.Join(mergeHash, []byte{})
	hash := sha256.Sum256(headers)

	//[32]byte -> []byte
	b.Hash = hash[:]
}

//序列化区块信息

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	//编码器
	encoder := gob.NewEncoder(&result)

	//编码
	encoder.Encode(b)
	return result.Bytes()
}

//反序列化解码区块信息

func UnSerialize(b []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(b))

	decoder.Decode(&block)

	return &block
}

//构建区块交易hash

func (b *Block) HashTransactions() []byte {
	var (
		txHashs [][]byte
		txHash  [32]byte
	)

	for _, tx := range b.Transactions {
		txHashs = append(txHashs, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashs, []byte{}))

	return txHash[:]
}
