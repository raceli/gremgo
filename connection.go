package gremgo

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type dialer interface {
	connect() error
	write([]byte) error
	read() ([]byte, error)
	close() error
	getAuth() *auth
}

/////
/*
WebSocket Connection
*/
/////

// Ws is the dialer for a WebSocket connection
type Ws struct {
	host string
	conn *websocket.Conn
	auth *auth
}

//Auth is the container for authentication data of dialer
type auth struct {
	username string
	password string
}

func (ws *Ws) connect() (err error) {
	d := websocket.Dialer{
		WriteBufferSize:  8192,
		ReadBufferSize:   8192,
		HandshakeTimeout: 5 * time.Second, // Timeout or else we'll hang forever and never fail on bad hosts.
	}
	ws.conn, _, err = d.Dial(ws.host, http.Header{})
	if err != nil {

		// As of 3.2.2 the URL has changed.
		// https://groups.google.com/forum/#!msg/gremlin-users/x4hiHsmTsHM/Xe4GcPtRCAAJ
		ws.host = ws.host + "/gremlin"
		ws.conn, _, err = d.Dial(ws.host, http.Header{})
	}
	return
}

func (ws *Ws) write(msg []byte) (err error) {
	err = ws.conn.WriteMessage(2, msg)
	return
}

func (ws *Ws) read() (msg []byte, err error) {
	_, msg, err = ws.conn.ReadMessage()
	return
}

func (ws *Ws) close() (err error) {
	err = ws.conn.Close()
	return
}

func (ws *Ws) getAuth() *auth{
	if ws.auth == nil {
		panic("You must create a Secure Dialer for authenticate with the server")
	}
	return ws.auth
}

/////

func (c *Client) writeWorker(errs chan error) { // writeWorker works on a loop and dispatches messages as soon as it recieves them
	for {
		select {
		case msg := <-c.requests:
			err := c.conn.write(msg)
			if err != nil {
				errs <- err
				c.Errored = true
				break
			}
		}
	}
}

func (c *Client) readWorker(errs chan error) { // readWorker works on a loop and sorts messages as soon as it recieves them
	for {
		msg, err := c.conn.read()
		if err != nil {
			errs <- err
			c.Errored = true
			break
		}
		if msg != nil {
			c.handleResponse(msg)
		}
	}
}
