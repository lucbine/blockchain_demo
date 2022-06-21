package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 //循环上限
)

//难度值

const targetBits = 10

//pow 结构

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

//创建工作量证明

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)

	//target 为1向左移位256-24
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

//开始执行

func (pow *ProofOfWork) Run() (int64, []byte) {
	var (
		hashInt     big.Int
		currentHash []byte
		nonce       = 0
	)

	fmt.Printf("Mining the block containing %s,maxNonce=%d\n", pow.block.Data, maxNonce)

	for nonce < maxNonce {
		currentHash = pow.getNonceHash(int64(nonce))
		hashInt.SetBytes(currentHash) //将hash换成大整数（用来对比的数据）

		//挖矿校验
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	fmt.Printf("\n\n")
	return int64(nonce), currentHash
}

func (pow *ProofOfWork) getNonceHash(nonce int64) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		Int2Hex(pow.block.Timestamp),
		Int2Hex(int64(targetBits)),
		Int2Hex(nonce),
	}, []byte{})

	hash := sha256.Sum256(data)

	fmt.Printf("\r%x --nonce:%d", hash, nonce)

	return hash[:]
}

// 校验区块的正确性

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	currentHash := pow.getNonceHash(pow.block.Nonce)

	hashInt.SetBytes(currentHash)

	return hashInt.Cmp(pow.target) == -1
}

//将int64 写入到[]byte

func Int2Hex(num int64) []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, num)
	return buff.Bytes()
}
