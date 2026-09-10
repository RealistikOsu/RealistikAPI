package websockets

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/RealistikOsu/RealistikAPI/common"
)

// SubscribeMultiMatches subscribes to receiving information from completed
// games in multiplayer matches.
func SubscribeMultiMatches(c *conn, message incomingMessage) {
	multiSubscriptionsMtx.Lock()
	var found bool
	for _, el := range multiSubscriptions {
		if el.ID == c.ID {
			found = true
			break
		}
	}
	// if it was not found, we need to add it
	if !found {
		multiSubscriptions = append(multiSubscriptions, c)
	}
	multiSubscriptionsMtx.Unlock()

	c.WriteJSON(TypeSubscribedToMultiMatches, nil)
}

var multiSubscriptions []*conn
var multiSubscriptionsMtx = new(sync.RWMutex)

func matchRetriever() {
	ps := red.Subscribe(context.Background(), "api:mp_complete_match")
	for {
		msg, err := ps.ReceiveMessage(context.Background())
		if err != nil {
			common.GenericError(err)
			return
		}
		go handleNewMultiGame(msg.Payload)
	}
}

func handleNewMultiGame(payload string) {
	defer catchPanic()
	multiSubscriptionsMtx.RLock()
	cp := make([]*conn, len(multiSubscriptions))
	copy(cp, multiSubscriptions)
	multiSubscriptionsMtx.RUnlock()

	for _, el := range cp {
		el.WriteJSON(TypeNewMatch, json.RawMessage(payload))
	}
}
