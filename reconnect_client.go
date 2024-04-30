package comfyUIclient

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReConnectClient struct {
	*Client
	quit chan interface{}
}

// NewReConnectClient ...
func NewReConnectClient(baseURL string, httpClient *http.Client) (*ReConnectClient, error) {

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("url.Parse: error: %w", err)
	}

	endPoint := NewEndPoint(baseURLParsed.Scheme, baseURLParsed.Hostname(), baseURLParsed.Port())

	c := &Client{
		ID:         uuid.NewString(),
		baseURL:    endPoint.String(),
		httpClient: httpClient,
		ch:         make(chan *WSMessage, 100),
	}

	endPoint.Protocol = "ws"
	if strings.HasPrefix(c.baseURL, "https") {
		endPoint.Protocol = "wss"
	}
	c.webSocket = NewWebSocketConnection(endPoint.String()+"/ws?clientId="+c.ID, 1, c)
	return &ReConnectClient{
		Client: c,
		quit:   make(chan interface{}),
	}, nil
}

func (rcc *ReConnectClient) Start() {
	go rcc.Run()
}

func (rcc *ReConnectClient) Run() {

	defer func() {
		close(rcc.ch)
		rcc.webSocket.SetIsConnected(false)
		if rcc.webSocket.Conn != nil {
			rcc.webSocket.Conn.Close()
		}
	}()
	for {
		select {
		case <-rcc.quit:
			fmt.Println("Exiting the handleWebSocket loop")
			return
		default:
			err := rcc.webSocket.Connect()
			if err != nil {
				fmt.Println("Error connecting to WebSocket:", err)
				time.Sleep(5 * time.Second) // 重连间隔
				continue
			}

			fmt.Println("Connected to WebSocket")
			fmt.Printf("Connected to WebSocket local:%s, remote:%s\n", rcc.webSocket.Conn.LocalAddr().String(), rcc.webSocket.Conn.RemoteAddr().String())

		loop:
			for {
				select {
				case <-rcc.quit:
					fmt.Println("Exiting the handleWebSocket loop")
					rcc.webSocket.Close()
					return
				default:
					_, message, err := rcc.webSocket.Conn.ReadMessage()
					if err != nil {
						fmt.Println("reading from WebSocket error: ", err)
						rcc.webSocket.Conn.Close()
						break loop
					}
					if err := rcc.Handle(string(message)); err != nil {
						log.Println("handle WebSocket error: ", err)
					}
				}
			}
		}
	}
}

func (rcc *ReConnectClient) Watch() chan *WSMessage {
	return rcc.GetTaskStatus()
}

func (rcc *ReConnectClient) Close() {
	rcc.quit <- true
}
