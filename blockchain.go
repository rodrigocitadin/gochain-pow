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
	"net"
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

func (bc *Blockchain) replaceChain(newBlocks []Block) bool {
	if len(newBlocks) <= len(bc.Blocks) {
		return false
	}

	if !isValidBlockchain(Blockchain{Blocks: newBlocks}) {
		return false
	}

	bc.Blocks = newBlocks
	fmt.Println("Blockchain replaced by the largest chain")
	return true
}

func requestBlockchain(address string) []Block {
	var newBlocks []Block

	for i := 0; i < 3; i++ { // retry 3 times
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println("Error connecting to peer:", address, "- tentativa", i+1)
			time.Sleep(2 * time.Second)
			continue
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET_CHAIN\n")
		decoder := json.NewDecoder(conn)

		if err := decoder.Decode(&newBlocks); err != nil {
			fmt.Println("Error decoding blockchain", address)
			return nil
		}
		return newBlocks
	}

	fmt.Println("Can't connect to", address)
	return nil
}

func calculateHash(index int, timestamp string, transactions []Transaction, prevHash string, nonce int) string {
	transactionsJSON, _ := json.Marshal(transactions)
	input := strconv.Itoa(index) + timestamp + string(transactionsJSON) + prevHash + strconv.Itoa(nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func mineBlock(prevBlock Block, transactions []Transaction) Block {
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
	return mineBlock(Block{Index: 0, Hash: "0"}, []Transaction{})
}

func isValidBlockchain(blockchain Blockchain) bool {
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

func (bc *Blockchain) minePendingTransactions(minerAddress string) {
	if len(bc.Mempool) == 0 {
		fmt.Println("Any transaction pending...")
		return
	}

	rewardTx := Transaction{Sender: "System", Receiver: minerAddress, Amount: miningReward}
	bc.Mempool = append(bc.Mempool, rewardTx)

	newBlock := mineBlock(bc.Blocks[len(bc.Blocks)-1], bc.Mempool)
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.Mempool = []Transaction{}

	fmt.Printf("Block %d mined from transactions mempool! Hash: %s\n", newBlock.Index, newBlock.Hash)
}
