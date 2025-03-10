
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

type Blockchain struct {
	Blocks []Block
}

func calculateHash(index int, timestamp, data, prevHash string, nonce int) string {
	input := strconv.Itoa(index) + timestamp + data + prevHash + strconv.Itoa(nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func mineBlock(prevBlock Block, data string, difficulty int) Block {
	timestamp := time.Now().String()
	index := prevBlock.Index + 1
	var nonce int
	var hash string
	target := strings.Repeat("0", difficulty)

	for {
		hash = calculateHash(index, timestamp, data, prevBlock.Hash, nonce)
		if strings.HasPrefix(hash, target) {
			break
		}
		nonce++
	}

	return Block{
		Index:     index,
		Timestamp: timestamp,
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Hash:      hash,
		Nonce:     nonce,
	}
}

func createGenesisBlock() Block {
	return mineBlock(Block{Index: 0, Hash: "0"}, "Genesis Block", 4)
}

func main() {
	blockchain := Blockchain{Blocks: []Block{createGenesisBlock()}}
	difficulty := 4

	for i := 1; i <= 5; i++ {
		newBlock := mineBlock(blockchain.Blocks[len(blockchain.Blocks)-1], fmt.Sprintf("Block %d", i), difficulty)
		blockchain.Blocks = append(blockchain.Blocks, newBlock)
		fmt.Printf("Bloco %d minerado! Hash: %s\n", newBlock.Index, newBlock.Hash)
	}
}
