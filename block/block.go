package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64  //时间
	Data          []byte //数据
	PrevBlockHash []byte //前块hash值
	Hash          []byte //当前块hash值
	Nonce         int64  //随机值
}

//创建创世纪块

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

//创建区块

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
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

	//将前一个hash、交易信息、时间戳 合并
	headers := bytes.Join([][]byte{tm, b.Data, b.PrevBlockHash}, []byte{})
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
