import net from 'net'
import bunyan from 'bunyan'
import mqtt from 'mqtt'

const log = bunyan.createLogger({
	name: (process.env['LOG_NAME'] || 'concava-connector-spul'),
	streams: [
		{ level: 'error', stream: process.stdout },
		{ level: 'info', path: '/logs/info.log' }
	]
})

const bigEndian     = (process.env['BIG_ENDIAN'] !== 'false' && process.env['BIG_ENDIAN'] !== '0')
const mqttHost      = (process.env['MQTT_HOST'] || 'unknown.host')
const headerSize    = (1 * process.env['HEADER_SIZE'] || 12)
const maxFrameSize  = (1 * process.env['MAX_FRAME_SIZE'] || 500)
const authToken     = (process.env['X_AUTH_TOKEN'] || 'unknown')
const timestampPort = 3333
const payloadPort   = 5555

// Timestamp server
var timestamps = net.createServer((socket) => {
	var timestamp = Math.round(Date.now() / 1000)
	var { remoteAddress, remotePort } = socket

	log.info({
		timestampPort, timestamp,
		remoteAddress, remotePort
	}, 'timestamp')

	var buf = new Buffer(4)
	buf[bigEndian ? 'writeUInt32BE' : 'writeUInt32LE'](timestamp)
	socket.end(buf)
})

timestamps.listen(timestampPort)
log.info('Timestamp server listening on ' + timestampPort)

// Payload server
var payloadServer = net.createServer((socket) => {
	socket.on('data', (buf) => {
		var timestamp = Math.round(Date.now() / 1000)
		var { remoteAddress, remotePort } = socket
		var deviceId  = buf.toString('hex', 0, 8)
		var numBlocks = buf.readInt8(8)
		var blockSize = buf.readInt8(9)

		log.info({
			payloadPort, timestamp,
			remoteAddress, remotePort,
			deviceId, numBlocks, blockSize
		}, 'data')

		var client = mqtt.connect('mqtt://' + mqttHost, {
			clientId: deviceId,
			password: authToken,
			connectTimeout: 3000, // ms
		})
		client.on('error', (err) => {
			log.error({
				payloadPort, timestamp,
				remoteAddress, remotePort,
				deviceId, err,
				buffer: buf.toString('hex')
			}, 'error')
		})
		client.on('connect', () => {
			for (let i = 0; i < numBlocks; i += 1) {
				let start = headerSize + i * blockSize
				let payload = buf.slice(start, start + blockSize)

				log.info({
					payloadPort, timestamp,
					remoteAddress, remotePort,
					deviceId,
					payload: payload.toString('hex')
				}, 'payload')

				client.publish('data', payload)
			}

			client.end()
			socket.end()
		})
	})
})

payloadServer.listen(payloadPort)
log.info('Payload server listening on ' + payloadPort)
