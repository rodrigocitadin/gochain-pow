package main

import (
	"fmt"
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

func main() {
       fmt.Println("Hi") 
}
