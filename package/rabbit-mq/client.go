package rabbitmq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	mu       sync.RWMutex
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &Client{
		conn:     conn,
		channels: make(map[string]*amqp.Channel),
	}, nil
}

func (c *Client) CreateChannel(name string) (*amqp.Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, exists := c.channels[name]; exists {
		return ch, nil
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	c.channels[name] = ch
	return ch, nil
}

func (c *Client) CloseChannel(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, exists := c.channels[name]; exists {
		if err := ch.Close(); err != nil {
			return err
		}
		delete(c.channels, name)
	}
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name, ch := range c.channels {
		ch.Close()
		delete(c.channels, name)
	}

	return c.conn.Close()
}

func (c *Client) DeclareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

func (c *Client) DeleteQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDelete(queueName, false, false, false)
	return err
}