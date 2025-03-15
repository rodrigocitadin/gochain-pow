package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

type Peer struct {
	Address string
}

type P2PNetwork struct {
	Peers []Peer
	mu    sync.Mutex
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

func (p *P2PNetwork) broadcast(data interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	message, _ := json.Marshal(data)
	for _, peer := range p.Peers {
		go sendMessage(peer.Address, message)
	}
}

func (p *P2PNetwork) addPeer(address string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Peers = append(p.Peers, Peer{Address: address})
}

func (p *P2PNetwork) startPeerServer(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting peer server on port", port, ":", err)
		return
	}
	defer ln.Close()

	fmt.Println("Peer server started on port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handlePeerConnection(conn)
	}
}

func handlePeerConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Received connection from", conn.RemoteAddr())
}

func StartPeers(ports []string) {
	p2pNetwork := P2PNetwork{}

	for _, port := range ports {
		p2pNetwork.addPeer("localhost:" + port)
		go p2pNetwork.startPeerServer(port)
	}
}

func listenForPeers() {
	for {
		time.Sleep(5 * time.Second)
		p2pNetwork.mu.Lock()
		for _, peer := range p2pNetwork.Peers {
			go func(peer Peer) {
				blocks := requestBlockchain(peer.Address)
				if blocks != nil {
					blockchain.replaceChain(blocks)
				}
			}(peer)
		}
		p2pNetwork.mu.Unlock()
	}
}

func sendMessage(address string, message []byte) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()
	conn.Write(message)
}

