package mssql

import (
	"fmt"
	"net"
	"time"
)

type timeoutConn struct {
	c             net.Conn
	timeout       time.Duration
	buf           *tdsBuffer
	packetPending bool
}

func NewTimeoutConn(conn net.Conn, timeout time.Duration) *timeoutConn {
	return &timeoutConn{
		c:       conn,
		timeout: timeout,
	}
}

func (c *timeoutConn) Read(b []byte) (n int, err error) {
	if c.buf != nil {
		if c.packetPending {
			c.packetPending = false
			err = c.buf.FinishPacket()
			if err != nil {
				err = fmt.Errorf("Cannot send handshake packet: %s", err.Error())
				return
			}
		}
		var packet uint8
		packet, err = c.buf.BeginRead()
		if err != nil {
			err = fmt.Errorf("Cannot read handshake packet: %s", err.Error())
			return
		}
		if packet != packPrelogin {
			err = fmt.Errorf("unexpected packet %d, expecting prelogin", packet)
			return
		}
		n, err = c.buf.Read(b)
		return
	}
	err = c.c.SetDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return
	}
	return c.c.Read(b)
}

func (c *timeoutConn) Write(b []byte) (n int, err error) {
	if c.buf != nil {
		if !c.packetPending {
			c.buf.BeginPacket(packPrelogin)
			c.packetPending = true
		}
		n, err = c.buf.Write(b)
		if err != nil {
			return
		}
		return
	}
	err = c.c.SetDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return
	}
	return c.c.Write(b)
}

func (c timeoutConn) Close() error {
	return c.c.Close()
}

func (c timeoutConn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c timeoutConn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c timeoutConn) SetDeadline(t time.Time) error {
	panic("Not implemented")
}

func (c timeoutConn) SetReadDeadline(t time.Time) error {
	panic("Not implemented")
}

func (c timeoutConn) SetWriteDeadline(t time.Time) error {
	panic("Not implemented")
}
