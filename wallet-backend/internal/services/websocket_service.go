package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// Client WebSocket客户端
type Client struct {
	ID      string
	UserID  uint64
	Socket  *websocket.Conn
	Send    chan []byte
	service *WebSocketService
}

// Message WebSocket消息
type Message struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	UserID uint64      `json:"user_id,omitempty"`
	Time   int64       `json:"time"`
}

// NewWebSocketService 创建新的WebSocket服务
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start 启动WebSocket服务
func (ws *WebSocketService) Start() {
	for {
		select {
		case client := <-ws.register:
			ws.mutex.Lock()
			ws.clients[client] = true
			ws.mutex.Unlock()
			log.Printf("Client %s connected", client.ID)

		case client := <-ws.unregister:
			ws.mutex.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.Send)
			}
			ws.mutex.Unlock()
			log.Printf("Client %s disconnected", client.ID)

		case message := <-ws.broadcast:
			ws.mutex.RLock()
			for client := range ws.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(ws.clients, client)
				}
			}
			ws.mutex.RUnlock()
		}
	}
}

// HandleWebSocket WebSocket处理器
func (ws *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint64) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 在生产环境中应该检查来源
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		ID:      fmt.Sprintf("client_%d_%d", userID, time.Now().Unix()),
		UserID:  userID,
		Socket:  conn,
		Send:    make(chan []byte, 256),
		service: ws,
	}

	ws.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.service.unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(512)
	c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Socket.SetPongHandler(func(string) error {
		c.Socket.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// 处理接收到的消息
		c.handleMessage(message)
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	// 根据消息类型处理
	switch msg.Type {
	case "ping":
		c.sendMessage(Message{
			Type: "pong",
			Time: time.Now().Unix(),
		})
	case "subscribe":
		// 处理订阅请求
		c.handleSubscribe(msg)
	case "unsubscribe":
		// 处理取消订阅请求
		c.handleUnsubscribe(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleSubscribe 处理订阅请求
func (c *Client) handleSubscribe(msg Message) {
	// 这里可以添加订阅逻辑，比如订阅特定地址的交易
	c.sendMessage(Message{
		Type: "subscribed",
		Data: map[string]interface{}{
			"message": "Successfully subscribed",
		},
		Time: time.Now().Unix(),
	})
}

// handleUnsubscribe 处理取消订阅请求
func (c *Client) handleUnsubscribe(msg Message) {
	c.sendMessage(Message{
		Type: "unsubscribed",
		Data: map[string]interface{}{
			"message": "Successfully unsubscribed",
		},
		Time: time.Now().Unix(),
	})
}

// sendMessage 发送消息给客户端
func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	c.Send <- data
}

// BroadcastToUser 向特定用户广播消息
func (ws *WebSocketService) BroadcastToUser(userID uint64, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for client := range ws.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(ws.clients, client)
			}
		}
	}
}

// BroadcastToAll 向所有客户端广播消息
func (ws *WebSocketService) BroadcastToAll(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	ws.broadcast <- data
}

// SendTransactionNotification 发送交易通知
func (ws *WebSocketService) SendTransactionNotification(userID uint64, transaction interface{}) {
	msg := Message{
		Type:   "transaction",
		Data:   transaction,
		UserID: userID,
		Time:   time.Now().Unix(),
	}
	ws.BroadcastToUser(userID, msg)
}

// SendBalanceUpdateNotification 发送余额更新通知
func (ws *WebSocketService) SendBalanceUpdateNotification(userID uint64, balance interface{}) {
	msg := Message{
		Type:   "balance_update",
		Data:   balance,
		UserID: userID,
		Time:   time.Now().Unix(),
	}
	ws.BroadcastToUser(userID, msg)
}

// SendDepositNotification 发送充值通知
func (ws *WebSocketService) SendDepositNotification(userID uint64, deposit interface{}) {
	msg := Message{
		Type:   "deposit",
		Data:   deposit,
		UserID: userID,
		Time:   time.Now().Unix(),
	}
	ws.BroadcastToUser(userID, msg)
}

// SendWithdrawNotification 发送提现通知
func (ws *WebSocketService) SendWithdrawNotification(userID uint64, withdraw interface{}) {
	msg := Message{
		Type:   "withdraw",
		Data:   withdraw,
		UserID: userID,
		Time:   time.Now().Unix(),
	}
	ws.BroadcastToUser(userID, msg)
}

// GetConnectedClientsCount 获取连接的客户端数量
func (ws *WebSocketService) GetConnectedClientsCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

// GetUserClientsCount 获取特定用户的客户端数量
func (ws *WebSocketService) GetUserClientsCount(userID uint64) int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	count := 0
	for client := range ws.clients {
		if client.UserID == userID {
			count++
		}
	}
	return count
}
