package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var BIG_ENDIAN, _ = strconv.ParseBool(os.Getenv("BIG_ENDIAN"))

const SPUL_URL = "spul_connector"
const SPUL_PORT = "5555"

const DEVICE_ID uint64 = 1234567890

var HEADER_SIZE, _ = strconv.ParseInt(os.Getenv("HEADER_SIZE"), 10, 64)

const MAX_NUM_BLOCKS = 4
const MAX_BLOCK_SIZE = 8

const MAX_FRAMES = 5

func main() {
	rand.Seed(time.Now().UnixNano())

	conn, err := net.Dial("tcp", SPUL_URL+":"+SPUL_PORT)
	if err != nil {
		fmt.Println("Error" + err.Error())
		return
	}
	defer conn.Close()
	
	iterations := int64(rand.Intn(MAX_FRAMES) + 1);
	
	for i := int64(0); i < iterations; i++ {
		numblocks := int64(rand.Intn(MAX_NUM_BLOCKS) + 1);
		blockSize := int64(rand.Intn(MAX_BLOCK_SIZE) + 1);
	
		sendBuffer := make([]byte, HEADER_SIZE+numblocks*blockSize)

		if BIG_ENDIAN {
			binary.BigEndian.PutUint64(sendBuffer, DEVICE_ID)
		} else {
			binary.LittleEndian.PutUint64(sendBuffer, DEVICE_ID)
		}

		sendBuffer[8] = uint8(numblocks)
		sendBuffer[9] = uint8(blockSize)
		sendBuffer[10] = 0
		sendBuffer[11] = 0
	
		for j := int64(0); j < (blockSize * numblocks); j++ {
			sendBuffer[j+HEADER_SIZE] = uint8(rand.Intn(255))
		}
		conn.Write(sendBuffer)
	}
}
