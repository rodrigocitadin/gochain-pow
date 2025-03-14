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
	rnd "math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
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

type Peer struct {
	Address string
}

type P2PNetwork struct {
	Peers []Peer
	mu    sync.Mutex
}

const (
	difficulty   = 4
	miningReward = 1
)

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

func (p *P2PNetwork) resolveConflicts(bc *Blockchain) {
	for _, peer := range p.Peers {
		go func(peer Peer) {
			blocks := requestBlockchain(peer.Address)
			if blocks != nil {
				bc.replaceChain(blocks)
			}
		}(peer)
	}
}

func requestBlockchain(address string) []Block {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return nil
	}
	defer conn.Close()

	fmt.Fprintf(conn, "GET_CHAIN\n")
	decoder := json.NewDecoder(conn)

	var newBlocks []Block
	if err := decoder.Decode(&newBlocks); err != nil {
		fmt.Println("Error decoding blockchain:", err)
		return nil
	}

	return newBlocks
}

func (p *P2PNetwork) addPeer(address string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Peers = append(p.Peers, Peer{Address: address})
}

func (p *P2PNetwork) broadcast(data interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message, _ := json.Marshal(data)
	for _, peer := range p.Peers {
		go sendMessage(peer.Address, message)
	}
}

func sendMessage(address string, message []byte) {
	conn, err := net.Dial("tpc", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()
	conn.Write(message)
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

func main() {
	blockchain := Blockchain{Blocks: []Block{createGenesisBlock()}}
	p2pNetwork := P2PNetwork{}

	p2pNetwork.addPeer("localhost:5001")
	p2pNetwork.addPeer("localhost:5002")
	p2pNetwork.addPeer("localhost:5003")

	go startMiner(&blockchain, "Miner")
	go simulateTransactions(&blockchain, &p2pNetwork)
	go startConflictResolver(&blockchain, &p2pNetwork)

	ln, err := net.Listen("tcp", ":5000")
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

		go handleConnection(conn, &blockchain)
	}
}

func startMiner(blockchain *Blockchain, minerAddress string) {
	for {
		time.Sleep(10 * time.Second)
		blockchain.minePendingTransactions(minerAddress)
	}
}

func simulateTransactions(blockchain *Blockchain, p2pNetwork *P2PNetwork) {
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

func startConflictResolver(blockchain *Blockchain, p2pNetwork *P2PNetwork) {
	for {
		time.Sleep(30 * time.Second)
		p2pNetwork.resolveConflicts(blockchain)
	}
}

func handleConnection(conn net.Conn, blockchain *Blockchain) {
	defer conn.Close()

	var request string
	fmt.Fscanf(conn, "%s", &request)

	if request == "GET_CHAIN" {
		encoder := json.NewEncoder(conn)
		encoder.Encode(blockchain.Blocks)
	}
}
