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

const NUM_BLOCKS = 4
const BLOCK_SIZE = 8

func main() {
	rand.Seed(time.Now().UnixNano())

	sendBuffer := make([]byte, HEADER_SIZE+BLOCK_SIZE*NUM_BLOCKS)

	if BIG_ENDIAN {
		binary.BigEndian.PutUint64(sendBuffer, DEVICE_ID)
	} else {
		binary.LittleEndian.PutUint64(sendBuffer, DEVICE_ID)
	}

	sendBuffer[8] = uint8(NUM_BLOCKS)
	sendBuffer[9] = uint8(BLOCK_SIZE)
	sendBuffer[10] = 0
	sendBuffer[11] = 0

	for i := int64(0); i < (BLOCK_SIZE * NUM_BLOCKS); i++ {
		sendBuffer[i+HEADER_SIZE] = uint8(rand.Intn(255))
	}

	conn, err := net.Dial("tcp", SPUL_URL+":"+SPUL_PORT)
	if err != nil {
		fmt.Println("Error" + err.Error())
		return
	}
	defer conn.Close()
	conn.Write(sendBuffer)
}
