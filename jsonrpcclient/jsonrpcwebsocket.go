package jsonrpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	Url              string
	connection       *websocket.Conn
	mutex            sync.Mutex
	IsConnected      bool
	Incoming         chan IncomingJsonRPCRequest
	outgoing         chan JsonRPCRequest
	responseHandlers map[string]chan JsonRPCResponse
	responseMutex    sync.Mutex
	nextId           int
	ctx              context.Context
	cancel           context.CancelFunc
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
		Url:              url,
		Incoming:         make(chan IncomingJsonRPCRequest),
		outgoing:         make(chan JsonRPCRequest),
		responseHandlers: make(map[string]chan JsonRPCResponse),
	}
}

func (c *Client) connect() error {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(c.Url, http.Header{})
	if err != nil {
		return err
	}
	c.connection = conn
	return nil
}

func (c *Client) Start() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	if err := c.connect(); err != nil {
		return err
	}
	go c.readMessages()
	go c.writeMessages()
	c.IsConnected = true
	c.emitConnected()
	return nil
}

func (c *Client) Stop(emit bool) error {
	if c.cancel != nil {
		c.cancel()
	}

	if c.connection != nil {
		c.connection.Close()
		c.IsConnected = false
		if emit {
			c.emitDisconnected()
		}
	}
	return nil
}

func (c *Client) readMessages() {
	for {
		c.connection.SetReadDeadline(time.Now().Add(time.Second * 1))

		_, message, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				// Normal closure, exit the loop
				return
			} else if e, ok := err.(net.Error); ok && e.Timeout() {
				// Read timeout occurred, check for shutdown signal and continue
				select {
				case <-c.ctx.Done():
					return
				default:
					continue
				}
			} else {
				// An actual error occurred, handle it
				log.Println("Error reading message:", err)
				c.IsConnected = false
				c.Stop(true)
				//c.Reconnect()
				return
			}
		}
		// handle a message
		var payload map[string]interface{}
		json.Unmarshal(message, &payload)
		if id, ok := payload["id"].(string); ok {
			response := JsonRPCResponse{ID: id, Result: payload["result"]}
			if err, ok := payload["error"].(map[string]interface{}); ok {
				response.Error = JsonRPCError{
					Code:    int(err["code"].(float64)),
					Message: err["message"].(string),
					Data:    err["data"],
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
				Method:  method,
				JsonRPC: "2.0",
			}
			if payload["params"] != nil {
				req.Params, _ = payload["params"].([]interface{})
			}
			c.Incoming <- req
		}
	}
}

func (c *Client) writeMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case message := <-c.outgoing:
			encoded, _ := json.Marshal(message)
			err := c.connection.WriteMessage(websocket.TextMessage, encoded)
			if err != nil {
				// Handle error (e.g., log it, trigger reconnection)
				fmt.Println("Error writing message:", err)
				return
			}
		}
	}
}

func (c *Client) Call(method string, params map[string]interface{}) (interface{}, error) {

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
	if response.Error != (JsonRPCError{}) {
		return nil, errors.New(response.Error.Message)
	}
	return response.Result, nil
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

func (c *Client) Upload(filename string, startPrint bool) {
	u, err := url.Parse(c.Url)
	if err != nil {
		panic(err)
	}
	u.Scheme = "http"
	u.Path = "/server/files/upload"
	file, _ := os.Open(filename)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("print", strconv.FormatBool(startPrint))
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	r, _ := http.NewRequest("POST", u.String(), body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		panic(err)
	}
}

func (c *Client) emitConnected() {
	req := IncomingJsonRPCRequest{JsonRPC: "2.0", Method: "_client_connected", Params: []interface{}{}}
	c.Incoming <- req
}

func (c *Client) emitDisconnected() {
	req := IncomingJsonRPCRequest{JsonRPC: "2.0", Method: "_client_disconnected", Params: []interface{}{}}
	c.Incoming <- req
}

func (c *Client) Reconnect() {
	timeout := 1
	for i := 1; i <= 10; i++ {
		if err := c.Start(); err == nil {
			break
		}
		time.Sleep(time.Duration(timeout * 1e9))
		timeout = timeout * 2
	}
}
