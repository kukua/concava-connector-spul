package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var BIG_ENDIAN, _ = strconv.ParseBool(os.Getenv("BIG_ENDIAN"))

var TIME_STAMP_PORT = "3333"
var SPUL_PORT = "5555"
var CONCAVA_URL = os.Getenv("CONCAVA_URL")

var HEADER_SIZE, _ = strconv.ParseInt(os.Getenv("HEADER_SIZE"), 10, 64)
var MAX_FRAME_SIZE, _ = strconv.ParseInt(os.Getenv("MAX_FRAME_SIZE"), 10, 64)

var X_AUTH_TOKEN = os.Getenv("X_AUTH_TOKEN")

var timeStampLog *os.File
var spulLog *os.File

func main() {
	timeStampLog, _ = os.Create("/var/log/spul_timestamps.txt")
	defer timeStampLog.Close()

	spulLog, _ = os.Create("/var/log/spul.txt")
	defer spulLog.Close()

	done := make(chan int)
	go setupTimeStampListener(done)
	go setupSPULListener(done)

	<-done
}

func setupTimeStampListener(done <-chan int) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+TIME_STAMP_PORT)
	if err != nil {
		fmt.Println("Timestamp socket listening error: " + err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Timestamp socket listening on: " + TIME_STAMP_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Timestamp accepting error: " + err.Error())
		} else {
			go handleTimeStampRequest(conn)
		}
	}
}

func handleTimeStampRequest(conn net.Conn) {
	fmt.Println(TIME_STAMP_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String())
	timeStampLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + "\r\n")

	bin := make([]byte, 4)
	if BIG_ENDIAN {
		binary.BigEndian.PutUint32(bin, uint32(time.Now().Unix()))
	} else {
		binary.LittleEndian.PutUint32(bin, uint32(time.Now().Unix()))
	}

	conn.Write(bin)
	conn.Close()
}

func setupSPULListener(done <-chan int) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+SPUL_PORT)
	if err != nil {
		fmt.Println("SPUL socket listening error: " + err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("SPUL socket listening on: " + SPUL_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("SPUL accepting error: " + err.Error())
		} else {
			go handleSPULRequest(conn)
		}
	}
}

func handleSPULRequest(conn net.Conn) {
	defer conn.Close()
	header := make([]byte, HEADER_SIZE)
	payload := make([]byte, MAX_FRAME_SIZE-HEADER_SIZE)

	headerSize, err := conn.Read(header)
	if err != nil {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header read error: " + err.Error())
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header read error" + err.Error() + "\r\n")
		return
	} else if int64(headerSize) != HEADER_SIZE {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header size error: " + err.Error())
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header size error" + err.Error() + "\r\n")
		return
	}

	payloadSize, err := conn.Read(payload)
	if err != nil {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload read error: " + err.Error())
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload read error" + err.Error() + "\r\n")
		return
	}

	var numBlocks int = int(header[8])
	var blockSize int = int(header[9])

	if (blockSize == 0) && (numBlocks == 1) {
		blockSize = payloadSize
	}

	if payloadSize != (numBlocks * blockSize) {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload size mismatch")
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload size mismatch\r\n")
		return
	}

	var deviceID uint64

	if BIG_ENDIAN {
		deviceID = binary.BigEndian.Uint64(header)
	} else {
		deviceID = binary.LittleEndian.Uint64(header)
	}

	fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + strconv.FormatUint(deviceID, 10) + ", " + strconv.FormatUint(uint64(numBlocks), 10) + ", " + strconv.FormatUint(uint64(blockSize), 10))
	spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + strconv.FormatUint(deviceID, 10) + ", " + strconv.FormatUint(uint64(numBlocks), 10) + ", " + strconv.FormatUint(uint64(blockSize), 10) + "\r\n")

	for i := 0; i < numBlocks; i++ {
		sendBuffer := make([]byte, blockSize)

		for j := 0; j < blockSize; j++ {
			sendBuffer[j] = payload[i*int(blockSize)+j]
		}

		go sendConcava(sendBuffer)
	}
}

func sendConcava(buffer []byte) {
	req, _ := http.NewRequest("POST", CONCAVA_URL, bytes.NewBuffer(buffer))
	req.Header.Set("X-Auth-Token", X_AUTH_TOKEN)
	req.Close = true

	var client http.Client
	client.Do(req)
}
