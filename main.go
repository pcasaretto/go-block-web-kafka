package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func main() {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "requests-0",
	})

	f := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestId := uuid.New()
		err := kafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:     []byte("Key-A"),
				Value:   []byte("Hello World!"),
				Headers: []kafka.Header{{Key: "request-id", Value: []byte(requestId.String())}},
			},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

	})
	http.HandleFunc("/request", f)
	http.ListenAndServe(":3001", nil)
	if err := kafkaWriter.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

func messageReader() {
	// make a new reader that consumes from topic-A, partition 0, at offset 42
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "responses",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(kafka.LastOffset)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
