package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	Sender    string
	Receiver  string
	Amount    int64
	Signature string
}

type Block struct {
	Index        int
	Timestamp    string
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Nonce        int
}

type Blockchain struct {
	Blocks  []Block
	Mempool []Transaction
}

const (
	difficulty = 4
)

func calculateHash(index int, timestamp string, transactions []Transaction, prevHash string, nonce int) string {
	transactionsJSON, _ := json.Marshal(transactions)
	input := strconv.Itoa(index) + timestamp + string(transactionsJSON) + prevHash + strconv.Itoa(nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func mineBlock(prevBlock Block, transactions []Transaction, difficulty int) Block {
	timestamp := time.Now().String()
	index := prevBlock.Index + 1
	var nonce int
	var hash string
	target := strings.Repeat("0", difficulty)

	for {
		hash = calculateHash(index, timestamp, transactions, prevBlock.Hash, nonce)
		if strings.HasPrefix(hash, target) {
			break
		}
		nonce++
	}

	return Block{
		Index:        index,
		Timestamp:    timestamp,
		Transactions: transactions,
		PrevHash:     prevBlock.Hash,
		Hash:         hash,
		Nonce:        nonce,
	}
}

func createGenesisBlock() Block {
	return mineBlock(Block{Index: 0, Hash: "0"}, []Transaction{}, 4)
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

		calculatedHash := calculateHash(currentBlock.Index, currentBlock.Timestamp, currentBlock.Transactions, currentBlock.PrevHash, currentBlock.Nonce)
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

func (bc *Blockchain) addTransaction(tx Transaction) {
	bc.Mempool = append(bc.Mempool, tx)
}

func generateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	if privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		panic(err)
	} else {
		return privateKey, &privateKey.PublicKey
	}
}

func signTransaction(tx Transaction, privateKey *ecdsa.PrivateKey) string {
	txHash := sha256.Sum256([]byte(tx.Sender + tx.Receiver + fmt.Sprintf("%d", tx.Amount)))
	if r, s, err := ecdsa.Sign(rand.Reader, privateKey, txHash[:]); err != nil {
		panic(err)
	} else {
		sig := r.Text(16) + ":" + s.Text(16)
		return sig
	}
}

func verifyTransaction(tx Transaction, publicKey *ecdsa.PublicKey) bool {
	parts := strings.Split(tx.Signature, ":")
	if len(parts) != 2 {
		return false
	}

	var r, s big.Int
	r.SetString(parts[0], 16)
	s.SetString(parts[1], 16)
	txHash := sha256.Sum256([]byte(tx.Sender + tx.Receiver + fmt.Sprintf("%d", tx.Amount)))

	return ecdsa.Verify(publicKey, txHash[:], &r, &s)
}

func (bc *Blockchain) minePendingTransactions(difficulty int) {
	for len(bc.Mempool) > 0 {
		newBlock := mineBlock(bc.Blocks[len(bc.Blocks)-1], bc.Mempool, difficulty)
		bc.Blocks = append(bc.Blocks, newBlock)
		bc.Mempool = bc.Mempool[1:len(bc.Mempool)]
		fmt.Printf("Block %d mined from mempool! Hash: %s\n", newBlock.Index, newBlock.Hash)
	}

	fmt.Println("Any transaction pending...")
}

func main() {
	blockchain := Blockchain{Blocks: []Block{createGenesisBlock()}}

	privateKey, _ := generateKeyPair()
	_, publicKey2 := generateKeyPair()

	tx := Transaction{Sender: "Alice", Receiver: "Bob", Amount: 10}
        tx.Signature = signTransaction(tx, privateKey)

        if verifyTransaction(tx, publicKey2) {
                fmt.Println("valid transaction")
                blockchain.addTransaction(tx)
        } else {
                fmt.Println("invalid transaction")
        }

	// blockchain.addTransaction(Transaction{Sender: "Alice", Receiver: "Bob", Amount: 10})
	// blockchain.addTransaction(Transaction{Sender: "Bob", Receiver: "Alice", Amount: 5})

	// blockchain.minePendingTransactions(difficulty)

	// fmt.Print("\n\nBlockchain:", blockchain, "\n\n")

	// if isValidBlockchain(blockchain, difficulty) {
	// 	fmt.Println("Valid blockchain")
	// } else {
	// 	fmt.Println("Invalid blockchain")
	// }
}
