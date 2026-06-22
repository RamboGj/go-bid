package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind = int

const (
	// Request
	PlaceBid MessageKind = iota

	// Success
	SuccessfullyPlacedBid

	// Errors
	FailedToPlaceBid
	InvalidJSON

	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message     string      `json:"message,omitempty"`
	Kind        MessageKind `json:"kind"`
	UserID      uuid.UUID   `json:"user_id,omitempty"`
	AmountCents int32       `json:"amount_cents,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsService *BidsService
}

func (ar *AuctionRoom) registerClient(client *Client) {
	slog.Info("New user Connected", "Client", client)
	ar.Clients[client.UserId] = client
}

func (ar *AuctionRoom) unregisterClient(client *Client) {
	slog.Info("User disconnected", "Client", client)
	delete(ar.Clients, client.UserId)
}

func (ar *AuctionRoom) broadcastMessage(message Message) {
	slog.Info("New message received", "RoomID", ar.Id, "message", message.Message, "user_id", message.UserID)

	switch message.Kind {
	case PlaceBid:
		bid, err := ar.BidsService.PlaceBid(ar.Context, ar.Id, message.UserID, message.AmountCents)
		if err != nil {
			if errors.Is(err, ErrBidIsTooLow) {
				if client, ok := ar.Clients[message.UserID]; ok {
					client.Send <- Message{
						Kind:    FailedToPlaceBid,
						Message: ErrBidIsTooLow.Error(),
					}
				}
			}
			return
		}

		if client, ok := ar.Clients[message.UserID]; ok {
			client.Send <- Message{
				Kind:        SuccessfullyPlacedBid,
				Message:     "Your bid was successfully placed.",
				AmountCents: bid.BidAmountCents,
				UserID:      message.UserID,
			}
		}

		for id, client := range ar.Clients {
			if id == message.UserID {
				continue
			}

			newBidMessage := Message{
				Kind:        NewBidPlaced,
				Message:     "New bid placed",
				AmountCents: bid.BidAmountCents,
				UserID:      message.UserID,
			}

			client.Send <- newBidMessage

		}
	case InvalidJSON:
		client, ok := ar.Clients[message.UserID]
		if !ok {
			slog.Info("Client not found in hashmap", "user_id", message.UserID)
			return
		}
		client.Send <- message
	}

}

func (ar *AuctionRoom) Run() {
	slog.Info("Auction has begun", "auctionRoomId", ar.Id)

	defer func() {
		close(ar.Broadcast)
		close(ar.Register)
		close(ar.Unregister)
	}()

	for {
		select {

		case client := <-ar.Register:
			ar.registerClient(client)
		case client := <-ar.Unregister:
			ar.unregisterClient(client)
		case message := <-ar.Broadcast:
			ar.broadcastMessage(message)
		case <-ar.Context.Done():
			slog.Info("Auction has ended.", "auctionID", ar.Id)
			for _, client := range ar.Clients {
				client.Send <- Message{
					Kind:    AuctionFinished,
					Message: "auction has been finished",
				}
			}
			return
		}
	}
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsService BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:          id,
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[uuid.UUID]*Client),
		Context:     ctx,
		BidsService: &bidsService,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	writeWait      = 10 * time.Second
	pingPeriod     = (readDeadline * 9 / 10)
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
		m.UserID = c.UserId

		err := c.Conn.ReadJSON(&m)
		if err != nil {
			var syntaxError *json.SyntaxError
			var unmarshalError *json.UnmarshalTypeError
			if errors.As(err, &syntaxError) || errors.As(err, &unmarshalError) {
				// The connection is still alive, the client just sent
				// something that isn't valid JSON. Report it and keep reading.
				c.Room.Broadcast <- Message{
					Kind:    InvalidJSON,
					Message: "this message should be a valid json",
					UserID:  m.UserID,
				}
				continue
			}

			// Any other error is connection-level: the socket has failed,
			// so we must stop reading or gorilla/websocket panics with
			// "repeated read on failed websocket connection".
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected Close error", "error", err)
			}
			return
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
					Message: "Closing websocket connection",
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
