import net from 'net'
import mqtt from 'mqtt'

const bigEndian     = process.env['BIG_ENDIAN'] !== 'false' && process.env['BIG_ENDIAN'] !== '0'
const mqttHost      = process.env['MQTT_HOST'] || 'unknown.host'
const headerSize    = 1 * process.env['HEADER_SIZE'] || 12
const maxFrameSize  = 1 * process.env['MAX_FRAME_SIZE'] || 500
const authToken     = process.env['X_AUTH_TOKEN'] || 'unknown'
const timestampPort = 3333
const payloadPort   = 5555

function log (...args) {
	console.log(...args)
}
function uts () {
	return Math.round(Date.now() / 1000)
}

// Timestamp server
var timestamps = net.createServer((socket) => {
	var timestamp = uts()

	log('%d: %d %s:%d', timestampPort, timestamp,
		socket.remoteAddress, socket.remotePort)

	var buf = new Buffer(4)
	buf[bigEndian ? 'writeUInt32BE' : 'writeUInt32LE'](timestamp)
	socket.end(buf)
})

timestamps.listen(timestampPort)
log('Timestamp server listening on', timestampPort)

// Payload server
var payloadServer = net.createServer((socket) => {
	socket.on('data', (buf) => {
		let timestamp = uts()
		let deviceId  = buf.toString('hex', 0, 8)
		let numBlocks = buf.readInt8(8)
		let blockSize = buf.readInt8(9)

		log('%d: %d %s:%d %s blocks: %d size: %d', payloadPort, timestamp,
			socket.remoteAddress, socket.remotePort, deviceId,
			numBlocks, blockSize)

		let client = mqtt.connect('mqtt://' + mqttHost, {
			clientId: deviceId,
			password: authToken,
			connectTimeout: 3000, // ms
		})
		client.on('error', (err) => {
			log('%d: %d %s:%d %s ERROR buffer: %s err: %s', payloadPort, timestamp,
				socket.remoteAddress, socket.remotePort, deviceId,
				buf.toString('hex'), err)
		})
		client.on('connect', () => {
			for (let i = 0; i < numBlocks; i += 1) {
				let start = headerSize + i * blockSize
				let payload = buf.slice(start, start + blockSize)

				log('%d: %d %s:%d %s SEND buffer: %s', payloadPort, timestamp,
					socket.remoteAddress, socket.remotePort, deviceId,
					payload.toString('hex'))

				client.publish('data', payload)
			}

			client.end()
			socket.end()
		})
	})
})

payloadServer.listen(payloadPort)
log('Payload server listening on', payloadPort)
