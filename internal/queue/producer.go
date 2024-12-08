package queue

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"product-management/internal/logging"
)

// connect initializes the RabbitMQ connection and channel
func connect() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	// Declare the image_processing queue without the x-dead-letter-exchange argument
	_, err = channel.QueueDeclare(
		"image_processing", // Queue name
		true,               // Durable
		false,              // Auto-delete
		false,              // Exclusive
		false,              // No-wait
		nil,                // No arguments (no dead-letter-exchange)
	)
	if err != nil {
		return nil, nil, err
	}

	return conn, channel, nil
}

// Publish sends a message to the RabbitMQ queue
func Publish(message interface{}) error {
	conn, channel, err := connect()
	if err != nil {
		logging.Logger.Error("Error connecting to RabbitMQ: ", err) // Log using Logrus
		return err
	}
	defer conn.Close()
	defer channel.Close()

	body, err := json.Marshal(message)
	if err != nil {
		logging.Logger.Error("Error marshaling message: ", err) // Log using Logrus
		return err
	}

	err = channel.Publish(
		"",                 // Exchange
		"image_processing", // Queue
		false,              // Mandatory
		false,              // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		logging.Logger.Error("Error publishing message to RabbitMQ: ", err) // Log using Logrus
		return err
	}

	logging.Logger.Info("Message published to RabbitMQ") // Log using Logrus
	return nil
}
