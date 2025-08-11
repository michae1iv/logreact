package kafka

import (
	"correlator/config"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaW struct {
	Conf     *config.KafkaProducerConfig
	Producer *kafka.Producer
	Status   string
	IncChn   chan []byte
}

func NewWriter() *KafkaW {
	return &KafkaW{
		IncChn: make(chan []byte),
	}
}

func (w KafkaW) GetChannel() chan []byte {
	return w.IncChn
}

func (w *KafkaW) LoadConfig(cfg interface{}) error {
	var err error

	kafkaCfg, ok := cfg.(*config.KafkaProducerConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected KafkaConsumerConfig")
	}

	if !kafkaCfg.Enable {
		return fmt.Errorf("kafka writer is disabled in conf file")
	}
	w.Conf = kafkaCfg

	// Setting up producer
	w.Producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": w.Conf.Brokers[0],
		"acks":              w.Conf.Acks,        // Ждём подтверждение от всех реплик (гарантия доставки)
		"retries":           w.Conf.Retries,     // Повторные попытки при ошибке
		"compression.type":  w.Conf.Сompression, // Опционально: сжатие
		"linger.ms":         w.Conf.LingerMS,    // Опционально: буферизация
		"batch.size":        w.Conf.BatchSize,   // Размер батча
	})
	if err != nil {
		return fmt.Errorf("failed to init producer: %w", err)
	}

	return nil
}

func (w KafkaW) Connect() error { // ! Заглушка
	return nil
}

func (w KafkaW) Resume() error { // ! Заглушка
	return nil
}

func (w KafkaW) Disconnect() {
	w.Producer.Close()
}

func (w KafkaW) Write(message []byte) error {
	err := w.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &w.Conf.Topic, Partition: kafka.PartitionAny},
		Key:            []byte("logreact"),
		Value:          message,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
