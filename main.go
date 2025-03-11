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

func isValidBlockchain(blockchain Blockchain, difficulty int) bool {
	for i := 1; i < len(blockchain.Blocks); i++ {
		prevBlock := blockchain.Blocks[i-1]
		currentBlock := blockchain.Blocks[i]

		if currentBlock.Index != prevBlock.Index+1 {
			return false
		}

		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}

		calculatedHash := calculateHash(currentBlock.Index, currentBlock.Timestamp, currentBlock.Data, currentBlock.PrevHash, currentBlock.Nonce)
		if currentBlock.Hash != calculatedHash {
			return false
		}

		target := strings.Repeat("0", difficulty)
		if !strings.HasPrefix(currentBlock.Hash, target) {
			return false
		}
	}
	return true
}

func main() {
	blockchain := Blockchain{Blocks: []Block{createGenesisBlock()}}
	difficulty := 4 

	for i := 1; i <= 5; i++ {
		newBlock := mineBlock(blockchain.Blocks[len(blockchain.Blocks)-1], fmt.Sprintf("Block %d", i), difficulty)
		blockchain.Blocks = append(blockchain.Blocks, newBlock)
		fmt.Printf("Block %d mined! Hash: %s\n", newBlock.Index, newBlock.Hash)
	}

	if isValidBlockchain(blockchain, difficulty) {
		fmt.Println("Valid blockchain")
	} else {
		fmt.Println("Invalid blockchain")
	}
}
