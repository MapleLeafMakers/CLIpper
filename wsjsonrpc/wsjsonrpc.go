package wsjsonrpc

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"reflect"
	"sync"
	"time"
)

const (
	maxReconnectAttempts = 5
	reconnectInterval    = 5 * time.Second
)

type JsonRPCRequest struct {
	Method  string      `json:"method"`
	ID      interface{} `json:"id"`
	JsonRPC string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
}

type JsonRPCError struct {
	Code    int
	Message string
	Data    interface{}
}

type JsonRPCResponse struct {
	ID     interface{}  `json:"id"`
	Result interface{}  `json:"result,omitempty"`
	Error  JsonRPCError `json:"error,omitempty"`
}

type RpcClient struct {
	Url    *url.URL
	conn   *websocket.Conn
	mu     sync.Mutex
	respMu sync.Mutex

	IsConnected       bool
	IsConnecting      bool
	shouldReconnect   bool
	reconnectAttempts int
	Incoming          chan JsonRPCRequest
	closeCh           chan struct{}
	responseChannels  map[interface{}]chan JsonRPCResponse
	nextId            float64
	isClosed          bool

	onConnect    func()
	onDisconnect func()
}

func NewWebSocketClient(serverUrl *url.URL) *RpcClient {
	return &RpcClient{
		Url:              serverUrl,
		shouldReconnect:  true,
		Incoming:         make(chan JsonRPCRequest),
		closeCh:          make(chan struct{}),
		responseChannels: make(map[interface{}]chan JsonRPCResponse),
	}
}

func (c *RpcClient) Connect() error {
	c.mu.Lock()
	c.IsConnecting = true
	c.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(c.Url.String(), nil)
	c.mu.Lock()
	if err != nil {
		c.IsConnecting = false
		c.mu.Unlock()
		return err
	}

	c.conn = conn
	c.IsConnected = true
	c.reconnectAttempts = 0
	c.mu.Unlock()

	go c.readPump()
	log.Println("Connected")
	if c.onConnect != nil {
		c.onConnect()
	}
	return nil
}

func (c *RpcClient) SetOnConnectFunc(f func()) {
	c.onConnect = f
}

func (c *RpcClient) SetOnDisconnectFunc(f func()) {
	c.onDisconnect = f
}

func (c *RpcClient) readPump() {
	defer func() {
		c.Disconnect()
		c.reconnect()
	}()

	for {
		select {
		case <-c.closeCh:
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			var req map[string]interface{}
			err = json.Unmarshal(msg, &req)
			if err != nil {
				panic("received invalid JsonRPC Request")
			}
			if _, ok := req["method"]; ok {
				// this is an incoming JsonRPCRequest
				var rpcReq JsonRPCRequest
				err = json.Unmarshal(msg, &rpcReq)
				if err != nil {
					panic("received invalid JsonRPCRequest")
				}

				c.Incoming <- rpcReq

			} else if id, ok := req["id"]; ok {
				// this is a response to a request we sent
				c.respMu.Lock()
				responseChan, ok := c.responseChannels[id]
				if ok {
					var resp JsonRPCResponse
					err = json.Unmarshal(msg, &resp)
					if err != nil {
						panic(err)
					}
					responseChan <- resp
					close(responseChan)
					delete(c.responseChannels, resp.ID)
				} else {
					log.Println("No response channel found for id", id, reflect.TypeOf(id))
				}
				c.respMu.Unlock()
			}

		}
	}
}

func (c *RpcClient) Call(method string, params map[string]interface{}) (interface{}, error) {
	responseChan := make(chan JsonRPCResponse, 1)
	c.respMu.Lock()
	c.nextId = c.nextId + 1
	c.responseChannels[c.nextId] = responseChan
	req := JsonRPCRequest{
		Method:  method,
		ID:      c.nextId,
		Params:  params,
		JsonRPC: "2.0",
	}
	c.respMu.Unlock()
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = c.WriteMessage(bytes)
	response := <-responseChan
	if response.Error != (JsonRPCError{}) {
		return nil, errors.New(response.Error.Message)
	}
	return response.Result, nil
}

func (c *RpcClient) WriteMessage(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsConnected {
		return errors.New("not connected")
	}

	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("write error: %v", err)
		c.Disconnect()
		return err
	}

	return nil
}

func (c *RpcClient) reconnect() {
	c.mu.Lock()
	if c.IsConnected || !c.shouldReconnect || c.reconnectAttempts >= maxReconnectAttempts {
		c.mu.Unlock()
		return
	}
	c.reconnectAttempts++
	c.mu.Unlock()

	time.Sleep(reconnectInterval)

	if err := c.Connect(); err != nil {
		log.Printf("reconnect failed: %v", err)
		go c.reconnect()
	}
}

func (c *RpcClient) Disconnect() {
	c.mu.Lock()
	if c.IsConnected {
		c.conn.Close()
		c.IsConnected = false
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}
	c.mu.Unlock()

}

func (c *RpcClient) Close() {

	c.mu.Lock()
	if c.isClosed {
		return
	} else {
		c.isClosed = true
	}
	c.mu.Unlock()

	c.shouldReconnect = false
	c.Disconnect()
	if c.Incoming != nil {
		close(c.Incoming)
	}
	if c.closeCh != nil {
		close(c.closeCh)
	}

}
