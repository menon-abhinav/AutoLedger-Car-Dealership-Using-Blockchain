package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"log"
)

const (
	MsgNewBlock                  = "newBlock"
	MsgConsensusResult           = "consensusResult"
	MsgBlockCreationConfirmation = "blockCreationConfirmation"
)

type Message struct {
	Type    string `json:"type"`    // "newBlock", "requestConsensus", etc.
	Content string `json:"content"` // JSON-encoded or base64-encoded content
}

func handleMessage(msg Message, ws *websocket.Conn, bc *Blockchain) {
	switch msg.Type {
	case MsgNewBlock:
		// Handle incoming new block message
		handleNewBlockMessage(msg.Content)
	case MsgConsensusResult:
		handleConsensusResult(msg, ws)
	case MsgBlockCreationConfirmation:
		// Handle incoming block creation confirmation message
		handleBlockCreationConfirmation(bc, msg.Content, ws)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func broadcastMessage(msg Message) {
	peers.Range(func(key, value interface{}) bool {
		ws, ok := key.(*websocket.Conn)
		if ok {
			err := ws.WriteJSON(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
				ws.Close()
				peers.Delete(ws)
			}
		}
		return true // Continue iteration
	})
}

func handleNewBlockMessage(block *Block) {
	new_msg := CreateBlockMessage(block)
	broadcastMessage(new_msg)
}

// handleMessage is assumed to be a part of the network message handler that calls specific functions based on message type.
func handleConsensusRequest(content string, ws *websocket.Conn) {
	// Simulating a consensus process. Normally, you'd parse the content to get the block data and validate it.
	block, err := DeserializeBlock([]byte(content)) // Assuming DeserializeBlock exists
	if err != nil {
		log.Printf("Failed to deserialize block: %v", err)
		return
	}

	// Here, initiate the actual consensus process, which might involve further communication and checks
	isConsensusReached := runConsensusAlgorithm(block)
	message := CreateConsensusResultMessage(isConsensusReached, content)

	sendMessage(ws, message)
}

func handleConsensusResult(msg Message, ws *websocket.Conn) {
	// Parse the content of the message to extract consensus result details
	var result struct {
		ConsensusReached bool   `json:"consensusReached"`
		Details          string `json:"details"`
	}
	if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
		log.Printf("Error parsing consensus result message: %v", err)
		return
	}

	// Process the consensus result
	if result.ConsensusReached {
		log.Printf("Consensus reached: %s", result.Details)
		// Handle the case when consensus is reached
	} else {
		log.Printf("Consensus not reached: %s", result.Details)
		// Handle the case when consensus is not reached
	}

	// Here you can perform any additional actions based on the consensus result,
	// such as updating the blockchain state or notifying other peers.
}

func handleBlockCreationConfirmation(bc *Blockchain, content string, ws *websocket.Conn) {
	// Decode the content from base64
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Printf("Error decoding base64 content: %v", err)
		return
	}

	block, err := DeserializeBlock(data)
	if err != nil {
		log.Printf("Failed to deserialize block: %v", err)
		return
	}

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if err := b.Put(block.Hash, block.Serialize()); err != nil {
			return err
		}

		if err := b.Put([]byte("l"), block.Hash); err != nil {
			return err
		}

		bc.tip = block.Hash
		return nil
	})
	if err != nil {
		log.Panic("Failed to update blockchain database: ", err)
	}
}

// CreateBlockMessage creates a network-ready message containing the serialized block.
func CreateBlockMessage(block *Block) Message {
	serializedBlock := block.Serialize()
	base64Block := base64.StdEncoding.EncodeToString(serializedBlock)

	return Message{
		Type:    "newBlock",
		Content: base64Block,
	}
}

func CreateConsensusResultMessage(isConsensusReached bool, details string) Message {
	// Construct the content to be a simple JSON object with consensus details
	content := fmt.Sprintf("{\"consensusReached\": %v, \"details\": \"%s\"}", isConsensusReached, details)
	return Message{
		Type:    MsgConsensusResult,
		Content: content,
	}
}

func CreateBlockCreationConfirmationMessage(block *Block) Message {
	// Encode the block hash to a hexadecimal string for readability and ease of transmission

	serializedBlock := block.Serialize()
	base64Block := base64.StdEncoding.EncodeToString(serializedBlock)

	return Message{
		Type:    MsgBlockCreationConfirmation,
		Content: base64Block,
	}
}

func sendMessage(ws *websocket.Conn, msg Message) error {
	return ws.WriteJSON(msg)
}

func runConsensusAlgorithm(block *Block) bool {
	// Implement your consensus logic here
	// For example, checking transactions, block hash, etc.
	return true // This is a placeholder
}
