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
	"os/exec"
	"time"
	// "runtime"


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
// go run main.go -l 10000 -secio
// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

type MinerDetails struct {
	UserId    	string
	NumBlocks 	int
	Faulty      int
	TimeToMine  int
	Age      	int
}

type rangeStake struct {
	Number int
}
var arraylist []rangeStake

var Miner MinerDetails

var age = 0

// Blockchain is a series of validated Blocks
var Blockchain []Block

var mutex = &sync.Mutex{}

var BasicHostId string
var globalHostId string

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
	// globalHostId = BasicHostId
	// fmt.Print("Global Host Id - " + globalHostId)


	// f, err := os.OpenFile("miner.txt", os.O_RDWR, 0644)
	// if err != nil {
 //    	panic(err)
	// }

	// defer f.Close()

	// if _, err = f.WriteString(BasicHostId); err != nil {
 //   		panic(err)
	// }


	// f1, err := os.OpenFile("minerdets.csv", os.O_RDWR, 0644)
	// if err != nil {
 //    	panic(err)
	// }

	// defer f1.Close()

	// if _, err = f1.WriteString("User ID\tNo. of Blocks Mined\tFaulty Transactions\tTime to Mine\tAge\n"); err != nil {
 //   		panic(err)
	// }


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

	log.Println("Got a new stream!")
	// fmt.Println("globalHostId - " + globalHostId)


	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

}

func readData(rw *bufio.ReadWriter) {
	var textToBeAdded string
	textToBeAdded = Miner.UserId + "," + strconv.Itoa(Miner.NumBlocks) + "," + strconv.Itoa(Miner.Faulty) + "," + strconv.Itoa(Miner.TimeToMine) + "," + strconv.Itoa(Miner.Age) + "\n" 
	mutex.Lock()
		f1, err := os.OpenFile("minerdets.csv",os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
    		panic(err)
		}

		defer f1.Close()

		if _, err = f1.WriteString(textToBeAdded); err != nil {
    		panic(err)
		}
			mutex.Unlock()

		//out, err := exec.Command("/home/nishantkr97/Downloads/Proof-of-AI---Blockchain/examples/p2p_2.0/MinerDetection.py").Output()
		// fmt.Println(out)
	

		// RUNNING PYTHON CODE


		cmd := exec.Command("/usr/bin/python3","/home/nishantkr97/Downloads/Proof-of-AI---Blockchain/examples/p2p_2.0/MinerDetection.py")
		out,err := cmd.CombinedOutput()

		if err != nil {
    		println("error is nibs " +err.Error()+" "+string(out))
    		return
		}

		fmt.Println(string(out))




    	// if err != nil {
     //    	fmt.Printf("%s", err)
    	// }

	    // fmt.Println("Command Successfully Executed")
    	// output := string(out[:])
    	// fmt.Println(output)

	for {
		age = age + 1
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}


		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			// Search element in arraylist 
			for i:=0;i<10;i++{
				if arraylist[i].Number == chain[len(chain)-1].BPM {
					arraylist[i].Number = -1
				}
			}
			fmt.Println(arraylist)



	file, err := os.Open("miner.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        globalHostId = scanner.Text()
        fmt.Println("The Miner is " + globalHostId)
    }


    fmt.Println("Miner details are: " + Miner.UserId + " num" + strconv.Itoa(Miner.NumBlocks))


    if err := scanner.Err(); err != nil {
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
	}
}

func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
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






		fmt.Print("> ")

		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		start := time.Now().UnixNano() / 1000000


		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

		if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
			mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}


		// Entered value
		valueEntered, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("You entered : " + strconv.Itoa(valueEntered))


		// Search element in arraylist 
		flag := 0
		for i:=0;i<10;i++{
			if arraylist[i].Number == valueEntered {
				arraylist[i].Number = -1
				flag = 1
			}
		}


		fmt.Println(arraylist)

		spew.Dump(Blockchain)
		end := time.Now().UnixNano() / 1000000
		timetaken := end - start

		fmt.Println("Time taken: " + strconv.Itoa(int(timetaken)))

		var textToBeAdded string
		if (flag == 1){
			textToBeAdded = BasicHostId + "\t" + "1\t" + "0\t" + strconv.Itoa(int(timetaken)) + "\t" + strconv.Itoa(age) + "\n"     
		} else {
			textToBeAdded = BasicHostId + "\t" + "0\t" + "0\t" + strconv.Itoa(int(timetaken)) + "\t" + strconv.Itoa(age) + "\n"     
		}
		

		age = 0


		f, err := os.OpenFile("data.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
    		panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(textToBeAdded); err != nil {
    		panic(err)
		}



		
		textToBeAdded = Miner.UserId + "," + strconv.Itoa(Miner.NumBlocks) + "," + strconv.Itoa(Miner.Faulty) + "," + strconv.Itoa(Miner.TimeToMine) + "," + strconv.Itoa(Miner.Age) + "\n" 
		mutex.Lock()
		f1, err := os.OpenFile("minerdets.csv",os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
    		panic(err)
		}

		defer f1.Close()

		if _, err = f1.WriteString(textToBeAdded); err != nil {
    		panic(err)
		}
			mutex.Unlock()

		if(globalHostId == BasicHostId){
			cmd := exec.Command("/usr/bin/python3","/home/nishantkr97/Downloads/Proof-of-AI---Blockchain/examples/p2p_2.0/MinerDetection.py")
			out,err := cmd.CombinedOutput()

			if err != nil {
	    		println("error is nibs " +err.Error()+" "+string(out))
	    		return
			}

			fmt.Println(string(out))
		}




		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))

		rw.Flush()
		mutex.Unlock()



	}

}

func main() {

	// f, err := os.Create("data.txt")
 //    if err != nil {
 //        fmt.Println(err)
 //       	return
 //    }

 //    l, err := f.WriteString("")
 //    if err != nil {
 //        fmt.Println(err)
 //        f.Close()
 //        return
 //    }
 //    fmt.Println(l, "bytes written successfully")
 //    err = f.Close()
 //    if err != nil {
 //        fmt.Println(err)
 //        return
 //    }

 //    fm, err := os.Create("miner.txt")
    // if err != nil {
    //     fmt.Println(err)
    //    	return
    // }
    // fmt.Println(fm, "Miner File Created")




	for i:=0;i<10;i++{
		newVal := rangeStake{}
		newVal = rangeStake{i}
		arraylist = append(arraylist, newVal)
	}
	fmt.Println(arraylist)

	// Uncomment later ---- originalArrayList := arraylist


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
		Miner.UserId = BasicHostId
    Miner.NumBlocks = 1
    Miner.Faulty = 99
    Miner.TimeToMine = 99
    Miner.Age = 1

  //   textToBeAdded := "User ID"+"\t" +"No. of Blocks Mined"+"\t"+"Faulty Transactions"+"\t"+"Time to Mine"+"\t"+"Age"+"\n"

  //   f, err := os.OpenFile("minerdets.csv", os.O_RDWR|os.O_WRONLY, 0600)
		// if err != nil {
  //   		panic(err)
		// }


		// if _, err = f.WriteString(textToBeAdded); err != nil {
  //   		panic(err)
		// }
		// defer f.Close()


		select {} // hang forever
		/**** This is where the listener code ends ****/
	} else {
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
		Miner.UserId = BasicHostId
	    Miner.NumBlocks = 99
	    Miner.Faulty = 0
	    Miner.TimeToMine = 1
	    Miner.Age = 99

		// Create a thread to read and write data.
		// fmt.Println("hi befre write")
		go readData(rw)
				// fmt.Println("hi between")

		go writeData(rw)
		
		// fmt.Println("hi after read")

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

	if(globalHostId != BasicHostId){
		Miner.Age = 1
		Miner.TimeToMine = 100
		Miner.Faulty = 100
		Miner.NumBlocks = 1

		// Miner.Age += 1
		// arraylist[newBlock.BPM].Number = newBlock.BPM
		fmt.Println("Terminal - " + BasicHostId + " - " + globalHostId)
		return false
	}else{
		Miner.NumBlocks += 100
	    Miner.Age = 100

	    // Miner.NumBlocks += 1
	    // Miner.Age = 0
	    return true

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