package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	// Requests
	PlaceBid MessageKind = iota

	// Ok/Success
	SuccessfullyPlacedBid

	// Errors
	FailedToPlaceBid
	InvalidJSON

	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	UserID  uuid.UUID   `json:"user_id,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
	Message string      `json:"message,omitempty"`
	Kind    MessageKind `json:"kind"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	ID         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsService BidsService
}

func (r *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user connected", "Client", c)
	r.Clients[c.UserID] = c
}

func (r *AuctionRoom) unregisterClient(c *Client) {
	slog.Info("user disconnected", "Client", c)
	delete(r.Clients, c.UserID)
}

func (r *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "Room ID", r.ID, "message", m.Message, "user_id", m.UserID)

	switch m.Kind {
	case PlaceBid:
		slog.Info("Kind certo")
		bid, err := r.BidsService.PlaceBid(r.Context, r.ID, m.UserID, m.Amount)
		if err != nil {
			slog.Error("Deu erro", "err:", err)
			if errors.Is(err, ErrBidIsTooLow) {
				slog.Error("Deu erro bid too low", "err:", err)
				if client, ok := r.Clients[m.UserID]; ok {
					client.Send <- Message{Kind: FailedToPlaceBid, Message: ErrBidIsTooLow.Error()}
				}
				slog.Error("User not okay :(", "clients", r.Clients, "user_id", m.UserID)
			}
			return
		}

		if client, ok := r.Clients[m.UserID]; ok {
			client.Send <- Message{Kind: SuccessfullyPlacedBid, Message: "Your bid was successfully placed."}
		}

		for _, client := range r.Clients {
			newBidMessage := Message{Kind: NewBidPlaced, Message: "A new bid was placed", Amount: bid.BidAmount}
			client.Send <- newBidMessage
		}
	case InvalidJSON:
		slog.Info("Kind invalid JSON")
		client, ok := r.Clients[m.UserID]
		if !ok {
			slog.Error("Client not found in hasmap", "user_id", m.UserID)
			return
		}
		client.Send <- m
	default: // New: Handle all other message kinds
		slog.Info("Kind default")
		for id, client := range r.Clients {
			if id == m.UserID {
				continue // Do not send the message back to the sender
			}
			client.Send <- m
		}
	}
}

func (r *AuctionRoom) Run() {
	slog.Info("NewAuctionRoom", "Room ID", r.ID)
	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
	}()

	for {
		select {
		case client := <-r.Register:
			r.registerClient(client)
		case client := <-r.Unregister:
			r.unregisterClient(client)
		case message := <-r.Broadcast:
			r.broadcastMessage(message)

		case <-r.Context.Done():
			slog.Info("Auction has ended.", "auctionID", r.ID)
			for _, client := range r.Clients {
				client.Send <- Message{Kind: AuctionFinished, Message: "auction has been finished"}
			}
			return
		}
	}
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, BidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		ID:          id,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		Context:     ctx,
		BidsService: BidsService,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserID uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserID: userID,
	}
}

const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	pingPeriod     = (readDeadline * 9) / 10 // 90% of readDeadline
	writeWait      = 10 * time.Second
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		var m Message
		m.UserID = c.UserID
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected close error", "error", err)
				return
			}
			c.Room.Broadcast <- Message{
				Kind:    InvalidJSON,
				Message: "This message should be valid JSON",
				UserID:  m.UserID,
			}
			continue
		}

		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{
					Kind:    websocket.CloseMessage,
					Message: "closing websocket connetction...",
				})
				return
			}
			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Room.Unregister <- c
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Unexpected write error", "error", err)
				return
			}
		}
	}
}
