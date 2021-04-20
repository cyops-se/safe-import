package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"

	"github.com/cyops-se/safe-import/usvc"
	"github.com/cyops-se/safe-import/usvc/examples/chatservice/types"
	"github.com/nats-io/nats.go"
)

func main() {
	nickname := flag.String("n", "guest", "The name you want to use in the chat")
	flag.Parse()

	broker := &usvc.UsvcBroker{}
	broker.Initialize()
	svc := usvc.CreateStub(broker, "chat", "examples", 1)

	if err := svc.SubscribeData("msg", msgHandler); err != nil {
		// fmt.Println("Failed to set up subscription:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		payload := types.ChatMessage{From: *nickname, Message: scanner.Text()}
		if _, err := svc.RequestMessage("msg", payload); err != nil {
			// fmt.Println("ERROR:", err)
		}
	}
}

func msgHandler(msg *nats.Msg) {
	payload := &types.ChatMessage{}
	json.Unmarshal(msg.Data, &payload)
	// fmt.Printf("%s says: %s\n", payload.From, payload.Message)
}
