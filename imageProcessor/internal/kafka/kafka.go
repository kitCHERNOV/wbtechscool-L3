package kafka

import "github.com/IBM/sarama"

func CreateProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_8_0_0 // Match your Kafka version
}
