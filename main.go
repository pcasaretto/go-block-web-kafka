package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

var port = flag.String("p", "3000", "PORT")

func main() {
	flag.Parse()

	ctx := context.Background()
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "requests-0",
	})

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "responses",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	kafkaReader.SetOffset(kafka.LastOffset)
	blocker := NewKafkaBlocker(kafkaReader)

	f := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestId := uuid.New()
		log.Println(requestId)
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
		select {
		case <-blocker.Block(requestId):
			w.WriteHeader(http.StatusOK)
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
		}
	})

	go blocker.Run(ctx)
	http.HandleFunc("/request", f)
	http.ListenAndServe(":"+*port, nil)
	if err := kafkaWriter.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	if err := kafkaReader.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}

}
