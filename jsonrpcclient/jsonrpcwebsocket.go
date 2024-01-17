package jsonrpcclient

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strconv"
	"sync"
)

type Client struct {
	url              string
	connection       *websocket.Conn
	mutex            sync.Mutex
	isConnected      bool
	Incoming         chan IncomingJsonRPCRequest
	outgoing         chan JsonRPCRequest
	responseHandlers map[string]chan JsonRPCResponse
	responseMutex    sync.Mutex
	nextId           int
}

type JsonRPCRequest struct {
	ID      string                 `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	JsonRPC string                 `json:"jsonrpc"`
}

type IncomingJsonRPCRequest struct {
	ID      string        `json:"id,omitempty"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	JsonRPC string        `json:"jsonrpc"`
}
type JsonRPCError struct {
	Code    int
	Message string
	Data    interface{}
}

type JsonRPCResponse struct {
	ID     string       `json:"id"`
	Result interface{}  `json:"result,omitempty"`
	Error  JsonRPCError `json:"error,omitempty"`
}

func NewClient(url string) *Client {
	return &Client{
		url:              url,
		Incoming:         make(chan IncomingJsonRPCRequest),
		outgoing:         make(chan JsonRPCRequest),
		responseHandlers: make(map[string]chan JsonRPCResponse),
	}
}

func (c *Client) Connect() error {
	var err error
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isConnected {
		return nil
	}

	u, _ := url.Parse(c.url)
	c.connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.isConnected = true
	go c.readMessages()
	go c.writeMessages()
	return nil
}

func (c *Client) readMessages() {
	defer close(c.Incoming)
	for {
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			c.isConnected = false
			c.Connect()
			return
		}
		//log.Println(string(message))
		var payload map[string]interface{}
		json.Unmarshal([]byte(message), &payload)
		if id, ok := payload["id"].(string); ok {
			response := JsonRPCResponse{ID: id, Result: payload["result"]}
			if error, ok := payload["error"].(map[string]interface{}); ok {
				log.Fatalf("%+v", error)
				response.Error = JsonRPCError{
					Code:    int(error["code"].(float64)),
					Message: error["message"].(string),
					Data:    error["data"],
				}
			}
			c.responseMutex.Lock()
			if responseChan, exists := c.responseHandlers[response.ID]; exists {
				delete(c.responseHandlers, response.ID)
				responseChan <- response
				close(responseChan)
			}
			c.responseMutex.Unlock()
		} else if method, ok := payload["method"].(string); ok {
			// incoming request

			req := IncomingJsonRPCRequest{
				Method: method,
				Params: payload["params"].([]interface{}),
			}
			c.Incoming <- req
		}
	}
}

func (c *Client) writeMessages() {
	for request := range c.outgoing {
		encoded, err := json.Marshal(request)
		if err != nil {
			log.Fatal("Encode Error", err)
		}
		err = c.connection.WriteMessage(websocket.TextMessage, encoded)
		if err != nil {
			log.Println("write error:", err)
			c.isConnected = false
			return
		}
	}
}

func (c *Client) Call(method string, params map[string]interface{}) JsonRPCResponse {

	responseChan := make(chan JsonRPCResponse, 1)
	c.responseMutex.Lock()
	c.nextId = c.nextId + 10
	id := strconv.Itoa(c.nextId)
	c.responseHandlers[strconv.Itoa(c.nextId)] = responseChan
	c.responseMutex.Unlock()
	if params == nil {
		params = map[string]interface{}{}
	}
	request := JsonRPCRequest{
		ID:      id,
		Method:  method,
		Params:  params,
		JsonRPC: "2.0",
	}

	c.outgoing <- request
	// Wait for the response
	response := <-responseChan

	return response
}

func (c *Client) Notify(method string, params map[string]interface{}) {
	if params == nil {
		params = map[string]interface{}{}
	}
	request := JsonRPCRequest{
		Method:  method,
		Params:  params,
		JsonRPC: "2.0",
	}
	c.outgoing <- request
}

func (c *Client) Close() {
	c.connection.Close()
	close(c.outgoing)
}
