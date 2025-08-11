package kafka

import (
	"correlator/config"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaR struct {
	Conf       *config.KafkaConsumerConfig
	Consumers  []*kafka.Consumer
	messageBuf chan []byte
	Status     string
	stopChan   chan struct{}
	once       sync.Once
}

func NewReader() *KafkaR {
	return &KafkaR{
		messageBuf: make(chan []byte, 10000),
		stopChan:   make(chan struct{}),
	}
}

func (r *KafkaR) LoadConfig(cfg interface{}) error {
	kafkaCfg, ok := cfg.(*config.KafkaConsumerConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected KafkaConsumerConfig")
	}
	if !kafkaCfg.Enable {
		return fmt.Errorf("kafka reader is disabled in conf file")
	}
	r.Conf = kafkaCfg

	for i := 0; i < 10; i++ {
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": r.Conf.Brokers[0],
			"group.id":          r.Conf.GroupID,
			"auto.offset.reset": r.Conf.AutoOffsetReset,
		})
		if err != nil {
			return fmt.Errorf("failed to create consumer #%d: %w", i, err)
		}
		r.Consumers = append(r.Consumers, consumer)
	}

	return nil
}

func (r *KafkaR) Connect() error {
	for i, consumer := range r.Consumers {
		if err := consumer.Subscribe(r.Conf.Topic, nil); err != nil {
			return fmt.Errorf("consumer #%d failed to subscribe: %w", i, err)
		}
	}

	r.once.Do(func() {
		for _, consumer := range r.Consumers {
			go r.pollMessages(consumer)
		}
	})

	return nil
}

func (r *KafkaR) pollMessages(consumer *kafka.Consumer) {
	for {
		select {
		case <-r.stopChan:
			return
		default:
			event := consumer.Poll(100)
			if event == nil {
				continue
			}
			switch e := event.(type) {
			case *kafka.Message:
				r.messageBuf <- e.Value
			case kafka.Error:
				fmt.Printf("Kafka error: %v\n", e)
			}
		}
	}
}

func (r *KafkaR) Resume() error {
	return r.Connect()
}

func (r *KafkaR) Disconnect() {
	close(r.stopChan)
	for _, c := range r.Consumers {
		c.Close()
	}
}

func (r *KafkaR) Read() ([][]byte, error) {
	var result [][]byte
	deadline := time.Now().Add(1000 * time.Millisecond)

	for len(result) < r.Conf.MaxPollRecords {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		select {
		case msg := <-r.messageBuf:
			result = append(result, msg)
		case <-time.After(remaining):
			break
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}
