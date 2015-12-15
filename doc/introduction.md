# Introduction

SPUL stands for __Sensor Protocol Ultra Light__ and is developed by [Sodaq](http://sodaq.com/). Spul is also a Dutch synonym for 'things'.

## Protocol for 2G/3G data connection

SPUL is built on top of the TCP protocol and uses TCP sockets for data communication. A SPUL TCP socket server will listen on incoming data packets, process these packets and forward the packets to [ConCaVa](https://github.com/kukua/concava). The server will listen on a configurable TCP port (`5555` by default). To avoid fragmenting of TCP frames the maximum size of the data packets is limited to 512 bytes.

The format of the TCP packet is defined as follows:

- 12 byte header block
- One or more data blocks

The header block contains the following bytes:

- 8 byte device ID
- 1 byte for the number of blocks
- 1 byte for the block length
- 2 bytes for the network (signal) quality

A byte block length value of `0` means the remaining bytes will be send as single data record (TCP packet size minus header size). In this case the number of blocks must be `0` aswell. This results in one block consisting of `TCP frame size - header size = 512 - 12 = 500 bytes`.

The TCP socket handler will receive a TCP frame, parse it into device ID & blocks, and forward it to ConCaVa. The device ID will passed along in the URL as a lowercase 16 character hexidecimal string (e.g. `http://concava.example/v1/sensorData/aabbccddeeff1234`). Per block a [PUT request](http://kukua.github.io/concava/latest/api/) will be made to ConCaVa with the block's data as binary body (content type `application/octet-stream`). Optionally an `Authorization` header is added for authentication in ConCaVa (e.g. `Authorization: Token <token>`).

## Data usage

For the Kukua weather stations a data record will be less than 20 bytes. A SPUL frame can hold at least 25 weather records. This allows us to send 20 records (3 minute samples) each hour in a single SPUL frame.

Mobile Data Operators often have a mimimum session of 1000 bytes. With the implementation we will remain well within this limit. With hourly uploads the monthly data charges will be: `24 * 1000 bytes * 31 = 727 kilobytes` which means a 1 MB/month bundle will be sufficient.
