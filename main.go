package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"time"
)

const (
	difficulty   = 4
	miningReward = 1
)

var (
	blockchain = Blockchain{Blocks: []Block{createGenesisBlock()}}
	serverPort = "5000"
	knownPeers = []string{"localhost:5001", "localhost:5002", "localhost:5003"}
)

func main() {
	go startMiner("Miner")
	go simulateTransactions()

        select {}
}

func startMiner(minerAddress string) {
	for {
		time.Sleep(10 * time.Second)
		blockchain.minePendingTransactions(minerAddress)
	}
}

func simulateTransactions() {
	users := map[string]*ecdsa.PrivateKey{}
	publicKeys := map[string]*ecdsa.PublicKey{}
	userNames := []string{"Alice", "Bob", "Charlie", "Dave"}

	for _, user := range userNames {
		privKey, pubKey := generateKeyPair()
		users[user] = privKey
		publicKeys[user] = pubKey
		fmt.Printf("Keys generated to %s\n", user)
	}

	for {
		time.Sleep(3 * time.Second)

		sender := userNames[rand.Intn(len(userNames))]
		receiver := userNames[rand.Intn(len(userNames))]
		if sender != receiver {
			amount := int64(rand.Intn(10) + 1)

			tx := Transaction{Sender: sender, Receiver: receiver, Amount: amount}
			tx.Signature = signTransaction(tx, users[sender])

			if verifyTransaction(tx, publicKeys[sender]) {
				blockchain.addTransaction(tx)
				fmt.Printf("New signed and validated transaction: %s → %s | Value: %d\n", sender, receiver, tx.Amount)
			} else {
				fmt.Println("Error: Invalid transaction detected")
			}
		}
	}
}
