package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	golog "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

type StakeBlock struct {
	User string
	Stake int
	NoOfUsers int
}

var flag2 int 

// Blockchain is a series of validated Blocks
var Blockchain []Block
var UserStakes []StakeBlock
var BasicHostId string
var choice int

var mutex = &sync.Mutex{}

var ValidatorID string

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID. 
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	BasicHostId = basicHost.ID().Pretty()

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

func handleStream(s net.Stream) {

	choice = 0

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	if(flag2 == 0){
		go readStake(rw,s)
		go writeStake(rw,s)
		go readData(rw)
		go writeData(rw)
		fmt.Println(" Broke out")
		go decideValidator(rw,s)
	}
	if flag2==1{
		go decideValidator(rw,s)
		go readData(rw)
		go writeData(rw)
	}
	
	// go readData(rw)
	// go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func decideValidator(rw *bufio.ReadWriter, s net.Stream){
	if flag2 ==0 {
		fmt.Println('0')
	}else{
	fmt.Println("\nLet's decide the winner...")

	ValidatorID = UserStakes[3].User
	fmt.Printf("\nWinner is : %s\n", ValidatorID)	

	// rw.Reader.Reset(bufio.NewReader(s))
	// rw.Writer.Reset(bufio.NewWriter(s))

	// mutex.Lock()
	if(BasicHostId == ValidatorID){
		go writeData(rw)
	}else {
		// fmt.Println("Going in....................")
		go readData(rw)
	}}
	// mutex.Unlock()

}

func readStake(rw *bufio.ReadWriter, s net.Stream) {
	flag := 0
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" || flag == 1{
			return
		}
		if str != "\n" {

			chain := make([]StakeBlock, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(UserStakes) {
				UserStakes = chain
				bytes, err := json.MarshalIndent(UserStakes, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				if(len(UserStakes) < 4){	
					fmt.Printf("\x1b[32m%s\x1b[0mEnter Token read: ", string(bytes))
				}else {
					fmt.Printf("\x1b[32m%s\x1b[0mAll the Validators have staked some amount! readstake!!\n", string(bytes))
					mutex.Unlock()
					// go decideValidator(rw,s)
					flag = 1
					flag2 = 1
					return
					//break
				}
			}
			mutex.Unlock()


		}
	}
}


func writeStake(rw *bufio.ReadWriter, s net.Stream) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(UserStakes)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		flag := 0
		for i := range UserStakes {
 		   if UserStakes[i].User == BasicHostId {
        		// Found!
        		flag = 1
    		}
		}

		if flag == 0 { 
			fmt.Print("Enter Token : ")
			sendData, err := stdReader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			sendData = strings.Replace(sendData, "\n", "", -1)
			bpm, err := strconv.Atoi(sendData)
			if err != nil {
				log.Fatal(err)
			}

			newStakeBlock := StakeBlock{BasicHostId, bpm, len(UserStakes) + 1}

	
			// Appending a new Stake Block
			mutex.Lock()
			UserStakes = append(UserStakes, newStakeBlock)
			mutex.Unlock()


			bytes, err := json.Marshal(UserStakes)
			if err != nil {
				log.Println(err)
			}

			spew.Dump(UserStakes)

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))


		
			rw.Flush()
			mutex.Unlock()
		}

		if(len(UserStakes) < 4){	
			// fmt.Printf("")
		}else{
			fmt.Printf("All the Validators have staked some amount!! writestake!\n")
			flag2 = 1
			return 
			// go decideValidator(rw,s)
			// break
		}		

	}
}


func readData(rw *bufio.ReadWriter) {
	fmt.Println("Galibaa000")
	//rw.Flush()


	for {
		if flag2==1 {
		str, err := rw.ReadString('\n')
		//fmt.Println(str)
		fmt.Println("Galibaa111")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Galibaa222")
		if str == "" {
			return
		}
		fmt.Println("Galibaa333")
		if str != "\n" && str[3] != 'U'{
			fmt.Printf("str -> %s\n", str)
			fmt.Println("Galibaa444")
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {
					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}}
}

func writeData(rw *bufio.ReadWriter) {
	//rw.Flush()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		if flag2 ==1{
		fmt.Print("Enter bpm : ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)


		// if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		// 	mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
		// 	mutex.Unlock()
		// }

		bytes2, err := json.Marshal(Blockchain)
		// fmt.Printf("Marshal Write Print-> %s\n", bytes)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(Blockchain)

		mutex.Lock()
		//rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes2)))

		//fmt.Println(rw.Flush())
		rw.Flush()
		mutex.Unlock()
		str, err := rw.ReadString('\n')
		fmt.Println("String is %s",str)
	}}

}

func main() {
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), ""}


	Blockchain = append(Blockchain, genesisBlock)

	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)
		fmt.Println("if target place of main before select\n")

		select {} // hang forever
		fmt.Println("this part shouldn't be printed\n")
		/**** This is where the listener code ends ****/
	} else {
		fmt.Println("entering else in main\n")
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// Create a thread to read and write User details
		s2, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		if flag2==0 {
		go readStake(rw,s2)
		go writeStake(rw,s2)
		go readData(rw)
		go writeData(rw)
			go decideValidator(rw,s2)
	}
		
		fmt.Println("Broke out")
		// go decideValidator(rw,s)



		// Create a thread to read and write data.
		// go writeData(rw)
		// go readData(rw)


		select {} // hang forever

	}
}

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}