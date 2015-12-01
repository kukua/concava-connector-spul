package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

var BIG_ENDIAN, _ = strconv.ParseBool(os.Getenv("BIG_ENDIAN"))
var TIME_STAMP_URL = "spul_connector"
var TIME_STAMP_PORT = "3333"

func main() {
	conn, err := net.Dial("tcp", TIME_STAMP_URL+":"+TIME_STAMP_PORT)
	if err != nil {
		fmt.Println("Error" + err.Error())
		return
	}

	timeStampBuff := make([]byte, 4)
	_, err = conn.Read(timeStampBuff)

	var timeStamp uint32

	if BIG_ENDIAN {
		timeStamp = binary.BigEndian.Uint32(timeStampBuff)
	} else {
		timeStamp = binary.LittleEndian.Uint32(timeStampBuff)
	}

	fmt.Println(timeStamp)
}
