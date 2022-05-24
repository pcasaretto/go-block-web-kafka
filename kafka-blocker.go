package main

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
	kafkaProtocol "github.com/segmentio/kafka-go/protocol"
)

type KafkaBlocker struct {
	*kafka.Reader
	mutex    *sync.RWMutex
	blockMap map[uuid.UUID]chan struct{}
}

func NewKafkaBlocker(reader *kafka.Reader) *KafkaBlocker {
	return &KafkaBlocker{
		Reader:   reader,
		mutex:    &sync.RWMutex{},
		blockMap: make(map[uuid.UUID]chan struct{}),
	}
}

func (kb *KafkaBlocker) Block(uuid uuid.UUID) <-chan struct{} {
	kb.mutex.Lock()
	ch := make(chan struct{})
	kb.blockMap[uuid] = ch
	kb.mutex.Unlock()
	return ch
}

func findHeader(key string, headers []kafkaProtocol.Header) (kafkaProtocol.Header, bool) {
	for _, h := range headers {
		if h.Key == key {
			return h, true
		}
	}
	return kafka.Header{}, false
}
func (kb *KafkaBlocker) Run(ctx context.Context) {
	for {
		m, err := kb.Reader.ReadMessage(ctx)
		if err != nil {
			continue
		}
		log.Println("got message")
		header, ok := findHeader("request-id", m.Headers)
		if !ok {
			continue
		}
		log.Println("found header", header)
		bytes := header.Value
		messageUUID, err := uuid.ParseBytes(bytes)
		if err != nil {
			continue
		}
		log.Println("got bytes", messageUUID)
		kb.mutex.RLock()
		log.Println("map", kb.blockMap[messageUUID])
		kb.mutex.RUnlock()
		if ch, ok := kb.blockMap[messageUUID]; ok {
			close(ch)
			kb.mutex.Lock()
			delete(kb.blockMap, messageUUID)
			kb.mutex.Unlock()
		}
	}
}
