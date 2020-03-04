package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	// "math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	// "io/ioutil"

	"github.com/davecgh/go-spew/spew"
	// "github.com/joho/godotenv"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	FileHash       string
	Hash      string
	PrevHash  string
	Validator string
	MerkelRoot string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block
var tempBlocks []Block

var WinnerIndex int = 0

var arraylist []int
var minerCapacity []int
// var winnerTimes []int


var lotteryPool []string


// candidateBlocks handles incoming blocks for validation
var candidateBlocks = make(chan Block)

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

var mutex = &sync.Mutex{}

// validators keeps track of open validators and balances
var validators = make(map[string]int)
var winnerTimes = make(map[string]int)
var finalWinnerTimes = make(map[string]int)
var count int = 0
var start time.Time
var cost int 

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// create genesis block
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), "Welcome", calculateBlockHash(genesisBlock), "", "", calculateMerkelHash(genesisBlock)}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	f, err := os.Create("data.txt")
    if err != nil {
        fmt.Println(err)
       	return
    }

    f1, err := os.OpenFile("data.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
    	panic(err)
	}

	defer f1.Close()

		

    count := 0
    size := 1024
    for{
    	if(count < size){
    		for i:=30;i<304;i++{
				// newVal := rangeStake{}
				// newVal = rangeStake{i}
				arraylist = append(arraylist, i)
				cost = i - 40

				if _, err = f1.WriteString(strconv.Itoa(i) + "\n"); err != nil {
    			panic(err)
				}
			}
    	}else{
    		break
    	}
    	count += 1
    }


    for i:=1;i<9;i++{
    	val := (i*100) % 13
    	minerCapacity = append(minerCapacity, val)
    }

    // fmt.Println(minerCapacity)

    err = f.Close()
    if err != nil {
        fmt.Println(err)
        return
    }

	// start TCP and serve TCP server
	// server, err := net.Listen("tcp", ":9000")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("HTTP Server Listening on port :9000")
	// defer server.Close()



	server, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

// pickWinner creates a lottery pool of validators and chooses the validator who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
// func pickWinner() {
func pickWinner() {
	time.Sleep(10 * time.Second)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	// fmt.Print("Miner Capacity - ")
 //    fmt.Println(minerCapacity)

	// lotteryPool := []string{}
	if len(temp) > 0 {

	OUTER:
		for _, block := range temp {
			// if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			// lock list of validators to prevent data race
			mutex.Lock()
			setValidators := validators
			mutex.Unlock()

			k, ok := setValidators[block.Validator]

			if ok {
				// if(k>lotteryMax){
				// 	lotteryMax = k
				// 	lotteryWinner = block.Validator
				// }

				k = 1

				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}

				
			}
		}

		if(count == 0){
			start := time.Now()
			fmt.Println(start)
		}


		fmt.Print("No. of times The Miner has been the Winner : ") 
		fmt.Println(finalWinnerTimes)
		lotteryWinner := lotteryPool[WinnerIndex%(len(lotteryPool))]
		WinnerIndex = WinnerIndex + 1
		

		max := -1
		win := ""
		for k, v := range validators{
			fmt.Println("k:", k, "v:", v)
			if(v > max){
				max = v
				win = k
			}
		}
		

		fmt.Println("Winner : " , win)
		lotteryWinner = win


		fmt.Print("Count = ")
		fmt.Println(count)
		count = count + 1
		if(count == 5){
			t := time.Now()
			elapsed := t.Sub(start)
			fmt.Print("Time Taken : ")
			fmt.Println(elapsed)
			count = 0
		}


	


		// add block of winner to blockcxhain and let all the other nodes know
		for _, block := range temp {
			if block.Validator == lotteryWinner {
				winnerTimes[lotteryWinner] = winnerTimes[lotteryWinner] + 1 
				finalWinnerTimes[lotteryWinner] =  finalWinnerTimes[lotteryWinner] + 1
				fmt.Println(winnerTimes)				
				if(winnerTimes[lotteryWinner] >= 2){
					validators[lotteryWinner] = validators[lotteryWinner] / 2 
					winnerTimes[lotteryWinner] = 0
					
				}
				if isTransactionValid(cost){
					validators[lotteryWinner] = validators[lotteryWinner] / 5
				}


				mutex.Lock()
				Blockchain = append(Blockchain, block)
				mutex.Unlock()
				for _ = range validators {
					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
				}
				break
			}
		}
	}

	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()
	// validator address
	var address string

	// allow user to allocate number of tokens to stake
	// the greater the number of tokens, the greater chance to forging a new block
	io.WriteString(conn, "Enter your capacity: ")
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now()
		address = calculateHash(t.String())
		validators[address] = balance
		winnerTimes[address] = 0
		fmt.Println(validators)
		break
	}

	io.WriteString(conn, "\nEnter a new FileName:")
	

	scanBPM := bufio.NewScanner(conn)

	go func() {
		for {
			// take in BPM from stdin and add it to blockchain after conducting necessary validation
			for scanBPM.Scan() {

				file, err := os.Open(scanBPM.Text())
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)

				var text string = ""
				var value string = ""

				for scanner.Scan() {             // internally, it advances token based on sperator
					// fmt.Println(scanner.Text()) 
					text = text + scanner.Text() // token in unicode-char
				}
				// fmt.Println(text)

				value = calculateHash(text)
				// fmt.Println(value)

				// newMerkelBlock := Block{}
				mutex.Lock()
				oldLastIndex := Blockchain[len(Blockchain)-1]
				mutex.Unlock()

				// create newBlock for consideration to be forged
				// newBlock, err := generateBlock(oldLastIndex, bpm, address)
				newBlock, err := generateBlock(oldLastIndex, value, address)

				if err != nil {
					log.Println(err)
					continue
				}


				if isBlockValid(newBlock, oldLastIndex) {
					candidateBlocks <- newBlock
				}


				io.WriteString(conn, "\nEnter a new FileName:")
			}
		}
	}()

	// simulate receiving broadcast
	for {
		time.Sleep(10*time.Second)
		mutex.Lock()
		output, err := json.MarshalIndent(Blockchain, "", "  ")
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}

}

// isBlockValid makes sure block is valid by checking index
// and comparing the hash of the previous block

// Transaction Verification 

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}


	return true
}


//calculateMerkelHash returns the hash of all block information
func calculateMerkelHash(block Block) string {
	record := string(block.Index) + string(block.FileHash) + block.PrevHash
	return calculateHash(record)
}


func isTransactionValid(cost int) bool {

	if(!(cost >= 30 && cost < 2000)){
		return false
	}

	return true
}

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.FileHash) + block.PrevHash
	return calculateHash(record)
}




// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, BPM string, address string) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.FileHash = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address
	newBlock.MerkelRoot = calculateMerkelHash(newBlock)
	return newBlock, nil
}



