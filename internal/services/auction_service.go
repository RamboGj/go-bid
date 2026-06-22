package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"

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

	// Info
	NewBidPlaced
	AuctionFinished
)

type Message struct {
	Message     string
	Kind        MessageKind
	UserID      uuid.UUID
	AmountCents int32
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
				Kind:    SuccessfullyPlacedBid,
				Message: "Your bid was successfully placed.",
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
			}

			client.Send <- newBidMessage

		}
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
