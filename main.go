package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	rnd "math/rand"
	"net"
	"time"
)

const (
	difficulty   = 4
	miningReward = 1
)

var (
	blockchain = Blockchain{Blocks: []Block{createGenesisBlock()}}
	p2pNetwork = P2PNetwork{}
	serverPort = "5000"
	knownPeers = []string{"localhost:5001", "localhost:5002", "localhost:5003"}
)

func main() {
	p2pNetwork.startPeerServer(serverPort)

	for _, peerAddr := range knownPeers {
		if peerAddr != "localhost:"+serverPort {
			p2pNetwork.addPeer(peerAddr)
		}
	}

	go startMiner("Miner")
	go startConflictResolver()
	go simulateTransactions()
	go listenForPeers()

	ln, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server started on port 5000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn)
	}
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

		sender := userNames[rnd.Intn(len(userNames))]
		receiver := userNames[rnd.Intn(len(userNames))]
		if sender != receiver {
			amount := int64(rnd.Intn(10) + 1)

			tx := Transaction{Sender: sender, Receiver: receiver, Amount: amount}
			tx.Signature = signTransaction(tx, users[sender])

			if verifyTransaction(tx, publicKeys[sender]) {
				blockchain.addTransaction(tx)
				p2pNetwork.broadcast(tx)
				fmt.Printf("New signed and validated transaction: %s → %s | Value: %d\n", sender, receiver, tx.Amount)
			} else {
				fmt.Println("Error: Invalid transaction detected")
			}
		}
	}
}

func startConflictResolver() {
	for {
		time.Sleep(30 * time.Second)
		p2pNetwork.resolveConflicts(&blockchain)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var request string
	fmt.Fscanf(conn, "%s", &request)

	if request == "GET_CHAIN" {
		encoder := json.NewEncoder(conn)
		encoder.Encode(blockchain.Blocks)
	}
}
