package accbroadcast

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	IncomingMessages chan interface{}

	conn   *net.UDPConn
	parser *messageParser
}

func NewClient(host string, port int) (*Client, error) {
	address := fmt.Sprintf("%v:%v", host, port)

	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	client := &Client{
		IncomingMessages: make(chan interface{}, 16),
		conn:             conn,
		parser:           newMessageParser(),
	}

	go client.readMessages()

	return client, nil
}

func (c *Client) Close() {
	if err := c.conn.Close(); err != nil {
		log.Printf("Cannot close UDP client connection: %v", err)
	}
}

func (c *Client) Register(displayName string, connectionPassword string, realtimeUpdateInterval time.Duration, commandPassword string) error {
	mb := newMessageBuilder()
	mb.WriteByte(outRegisterCommandApplication)
	mb.WriteByte(protocolVersion)

	mb.WriteString(displayName)
	mb.WriteString(connectionPassword)
	mb.WriteUint32(uint32(realtimeUpdateInterval / time.Millisecond))
	mb.WriteString(commandPassword)

	return c.sendMessage(mb.Bytes())
}

func (c *Client) Unregister(connectionId uint32) error {
	mb := newMessageBuilder()
	mb.WriteByte(outUnregisterCommandApplication)
	mb.WriteUint32(connectionId)

	return c.sendMessage(mb.Bytes())
}

func (c *Client) RequestEntryList(connectionId uint32) error {
	mb := newMessageBuilder()
	mb.WriteByte(outRequestEntryList)
	mb.WriteUint32(connectionId)

	return c.sendMessage(mb.Bytes())
}

func (c *Client) readMessages() {
	buf := make([]byte, 1024)

	for {
		size, _, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP client connection: %v", err)
			log.Printf("Stopping further message processing on UDP client connection")
			return
		}

		msg, err := c.parser.Parse(buf[0:size])
		if err != nil {
			log.Printf("Cannot parse message: %v", err)
		} else {
			c.IncomingMessages <- msg
		}
	}
}

func (c *Client) sendMessage(msg []byte) error {
	n, err := c.conn.Write(msg)
	if err == nil && n != len(msg) {
		return fmt.Errorf("wrote %v bytes out of %v", n, len(msg))
	}
	return err
}
