package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"be-lms/config"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 65536
)

// SafeConn is a thread-safe wrapper around websocket.Conn
type SafeConn struct {
	conn  *websocket.Conn
	mu    sync.Mutex
	close sync.Once
}

// NewSafeConn creates a new SafeConn
func NewSafeConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{conn: conn}
}

// WriteMessage writes a message with mutex protection
func (sc *SafeConn) WriteMessage(messageType int, data []byte) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteMessage(messageType, data)
}

// WriteJSON writes JSON with mutex protection
func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteJSON(v)
}

// SetWriteDeadline sets write deadline with mutex protection
func (sc *SafeConn) SetWriteDeadline(t time.Time) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets read deadline
func (sc *SafeConn) SetReadDeadline(t time.Time) error {
	return sc.conn.SetReadDeadline(t)
}

// SetReadLimit sets read limit
func (sc *SafeConn) SetReadLimit(limit int64) {
	sc.conn.SetReadLimit(limit)
}

// SetPongHandler sets pong handler
func (sc *SafeConn) SetPongHandler(h func(appData string) error) {
	sc.conn.SetPongHandler(h)
}

// ReadMessage reads a message
func (sc *SafeConn) ReadMessage() (messageType int, p []byte, err error) {
	return sc.conn.ReadMessage()
}

// Close closes the connection
func (sc *SafeConn) Close() error {
	var err error
	sc.close.Do(func() {
		sc.mu.Lock()
		err = sc.conn.Close()
		sc.mu.Unlock()
	})
	return err
}

// MessageHandler handles incoming messages from clients
type MessageHandler func(client *Client, envelope *WSEnvelope) error

// ClientReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump(messageHandler MessageHandler) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				config.Log.Warnf("⚠️ [Client] Read error: user_id=%d, course_id=%d, error=%v", c.UserID, c.CourseID, err)
			}
			break
		}

		// Parse the message envelope
		envelope, err := ParseEnvelope(message)
		if err != nil {
			config.Log.Warnf("⚠️ [Client] Invalid message format: user_id=%d, course_id=%d, error=%v", c.UserID, c.CourseID, err)
			// Send error to client
			errorEvent := NewErrorEvent("Invalid message format", "INVALID_FORMAT")
			if data, err := json.Marshal(errorEvent); err == nil {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.Conn.WriteMessage(websocket.TextMessage, data)
			}
			continue
		}

		// Handle the message
		if messageHandler != nil {
			if err := messageHandler(c, envelope); err != nil {
				config.Log.Errorf("❌ [Client] Handler error: user_id=%d, course_id=%d, type=%s, error=%v", c.UserID, c.CourseID, envelope.Type, err)
				errorEvent := NewErrorEvent(err.Error(), "HANDLER_ERROR")
				if data, err := json.Marshal(errorEvent); err == nil {
					c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
					c.Conn.WriteMessage(websocket.TextMessage, data)
				}
			}
		}
	}
}

// ClientWritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// Close Send channel only here, when WritePump exits
		// This is safe because no one else will try to send after client.Close() marks done
		close(c.Send)
		c.Conn.Close()
	}()

	for {
		select {
		case <-c.done:
			// Client marked as closed, exit gracefully
			return
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Send channel closed (shouldn't happen normally, but handle gracefully)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			// Check if client is closed before sending ping
			select {
			case <-c.done:
				return
			default:
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendEvent sends an event to the client
func (c *Client) SendEvent(event WSEnvelope) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if c.TrySend(data) {
		return nil
	}
	return ErrClientBufferFull
}

