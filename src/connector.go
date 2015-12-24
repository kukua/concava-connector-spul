package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	
	frameCount := 0
	
	for { 
		headerSize, err := conn.Read(header)
		if err != nil {
			if (frameCount == 0) {
				fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header read error: " + err.Error())
				spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header read error: " + err.Error() + "\r\n")
			}
			return
		} else if int64(headerSize) != HEADER_SIZE {
			fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header size error: " + err.Error())
			spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL header size error: " + err.Error() + "\r\n")
			return
		}
		
		frameCount++
		var numBlocks int = int(header[8])
		var blockSize int = int(header[9])
		var payloadSize int = numBlocks * blockSize;

		payload := make([]byte, payloadSize)
		
		_ , err = conn.Read(payload)
		if err != nil {
			fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload read error: " + err.Error())
			spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "SPUL payload read error" + err.Error() + "\r\n")
			return
		}

		var deviceID uint64

		if BIG_ENDIAN {
			deviceID = binary.BigEndian.Uint64(header)
		} else {
			deviceID = binary.LittleEndian.Uint64(header)
		}

		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", Frame: " + strconv.FormatUint(uint64(frameCount), 10) + ", " + strconv.FormatUint(deviceID, 16) + ", " + strconv.FormatUint(uint64(numBlocks), 10) + ", " + strconv.FormatUint(uint64(blockSize), 10))
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + strconv.FormatUint(deviceID, 16) + ", " + strconv.FormatUint(uint64(numBlocks), 10) + ", " + strconv.FormatUint(uint64(blockSize), 10) + "\r\n")

		for i := 0; i < numBlocks; i++ {
			sendBuffer := make([]byte, blockSize)

			for j := 0; j < blockSize; j++ {
				sendBuffer[j] = payload[i*int(blockSize)+j]
			}

			hexBuffer := hex.EncodeToString(sendBuffer)
			fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + strconv.FormatUint(deviceID, 16) + ", buffer: " + hexBuffer)
			spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + strconv.FormatUint(deviceID, 16) + ", buffer: " + hexBuffer + "\r\n")

			go sendConcava(deviceID, sendBuffer, conn)
		}
	}
}

func sendConcava(deviceID uint64, buffer []byte, conn net.Conn) {
	req, err := http.NewRequest("PUT", fmt.Sprintf(CONCAVA_URL, strings.ToLower(fmt.Sprintf("%016X", deviceID))), bytes.NewBuffer(buffer))
	if err != nil {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa NewRequest error: " + err.Error())
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa NewRequest error: " + err.Error() + "\r\n")
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Token "+X_AUTH_TOKEN)
	req.Close = true

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa PUT error: " + err.Error())
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa PUT error: " + err.Error() + "\r\n")
		return
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(SPUL_PORT + ": " + strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa error response: " + strconv.Itoa(resp.StatusCode) + " " + string(body))
		spulLog.WriteString(strconv.FormatInt(time.Now().Unix(), 10) + ", " + conn.RemoteAddr().String() + ", " + "ConCaVa error response: " + strconv.Itoa(resp.StatusCode) + " " + string(body) + "\r\n")
		return
	}
}
