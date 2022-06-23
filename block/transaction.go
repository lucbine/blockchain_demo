package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

//utxo 模型（unspent transaction outputs）

const subsidy = 10

//交易输入结构

type TXInput struct {
	TxId     []byte //引用交易ID
	VoutIdx  int    //引用的交易输出ID
	FromAddr string //输入方地址
}

//交易输出结构

type TXOutput struct {
	Value  int    //金额
	ToAddr string //输出方地址
}

//交易结构

type Transaction struct {
	ID   []byte     //交易ID
	Vin  []TXInput  //交易输入项
	Vout []TXOutput //交易输出项
}

// 将交易信息转换为hash 并设置ID

func (tx *Transaction) SetID() {
	//https://blog.csdn.net/idwtwt/article/details/80400314
	var buff bytes.Buffer
	var hash [32]byte

	//https://blog.csdn.net/weixin_42117918/article/details/105864520
	encoder := gob.NewEncoder(&buff)

	encoder.Encode(*tx)

	hash = sha256.Sum256(buff.Bytes())

	tx.ID = hash[:]
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	//创建一个输入项
	txInput := TXInput{[]byte{}, -1, data}

	//创建输出项
	txOutput := TXOutput{subsidy, to}

	tx := Transaction{nil, []TXInput{txInput}, []TXOutput{txOutput}}
	return &tx
}

//判断交易是为casebase 交易

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && tx.Vin[0].VoutIdx == -1 && len(tx.Vin[0].TxId) == 0
}

// 判断输入是否可以被某个账户使用

func (in *TXInput) CanUnlockOutputWith(address string) bool {
	return in.FromAddr == address
}

//判断某个输出是否可以被某个账号使用

func (out *TXOutput) CanBeUnlockedWith(address string) bool {
	return out.ToAddr == address
}

//创建普通的交易

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var (
		inputs  []TXInput
		outputs []TXOutput
	)

	//最小utxo
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	fmt.Println("-----", from, acc)

	if acc < amount {
		log.Panicln("ERROR: Not enough founds")
	}

	// 构建输入项

	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	//构建输出项
	outputs = append(outputs, TXOutput{amount, to})

	//输入值过大
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	//交易生成
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
