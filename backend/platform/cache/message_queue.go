/*
Hệ thống Redis-based message queue trong mã của bạn rất linh hoạt và có thể được sử dụng trong các trường hợp sau:
 1. Xử lý công việc bất đồng bộ (ví dụ: gửi email, thông báo, tính toán phức tạp).
 2. Quản lý retry và dead letter queue cho các công việc không thành công.
 4. Pub/Sub để xây dựng các ứng dụng thời gian thực hoặc hệ thống thông báo.
 3. Xử lý message tạm hoãn với Redis.
*/
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageQueue represents a Redis-based message queue
type MessageQueue struct {
	client *RedisClient
	ctx    context.Context
}

// NewMessageQueue creates a new message queue instance
func NewMessageQueue(ctx context.Context) (*MessageQueue, error) {
	client, err := NewRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	return &MessageQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

// Message represents a queue message
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
	Attempts  int                    `json:"attempts"`
	MaxRetry  int                    `json:"max_retry"`
}

// QueueOptions represents queue configuration
type QueueOptions struct {
	MaxRetry          int
	RetryDelay        time.Duration
	VisibilityTimeout time.Duration
}

// DefaultQueueOptions returns default queue options
func DefaultQueueOptions() *QueueOptions {
	return &QueueOptions{
		MaxRetry:          3,
		RetryDelay:        30 * time.Second,
		VisibilityTimeout: 5 * time.Minute,
	}
}

// Enqueue adds a message to the queue
func (mq *MessageQueue) Enqueue(queueName string, msgType string, payload map[string]interface{}, opts *QueueOptions) error {
	if opts == nil {
		opts = DefaultQueueOptions()
	}

	message := Message{
		ID:        generateID(),
		Type:      msgType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Attempts:  0,
		MaxRetry:  opts.MaxRetry,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return mq.client.Client.LPush(mq.ctx, queueName, data).Err()
}

// Dequeue retrieves a message from the queue
func (mq *MessageQueue) Dequeue(queueName string) (*Message, error) {
	// Use blocking pop operation
	result, err := mq.client.Client.BRPop(mq.ctx, 10*time.Second, queueName).Result()
	if err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue response")
	}

	var message Message
	err = json.Unmarshal([]byte(result[1]), &message)
	return &message, err
}

// DequeueNonBlocking retrieves a message from the queue without blocking
func (mq *MessageQueue) DequeueNonBlocking(queueName string) (*Message, error) {
	result, err := mq.client.Client.RPop(mq.ctx, queueName).Result()
	if err != nil {
		return nil, err
	}

	var message Message
	err = json.Unmarshal([]byte(result), &message)
	return &message, err
}

// RequeueFailed re-queues a failed message with incremented attempts
func (mq *MessageQueue) RequeueFailed(queueName string, message *Message, delay time.Duration) error {
	message.Attempts++

	if message.Attempts >= message.MaxRetry {
		// Move to dead letter queue
		return mq.EnqueueDeadLetter(queueName, message)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Add delay before re-queuing
	delayedQueueName := fmt.Sprintf("%s:delayed", queueName)
	score := float64(time.Now().Add(delay).Unix())

	return mq.client.Client.ZAdd(mq.ctx, delayedQueueName, redis.Z{
		Score:  score,
		Member: data,
	}).Err()
}

// EnqueueDeadLetter moves a message to the dead letter queue
func (mq *MessageQueue) EnqueueDeadLetter(queueName string, message *Message) error {
	deadLetterQueue := fmt.Sprintf("%s:dead", queueName)
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return mq.client.Client.LPush(mq.ctx, deadLetterQueue, data).Err()
}

// ProcessDelayedMessages moves delayed messages back to the main queue when ready
func (mq *MessageQueue) ProcessDelayedMessages(queueName string) error {
	delayedQueueName := fmt.Sprintf("%s:delayed", queueName)
	now := float64(time.Now().Unix())

	// Get messages ready to be processed
	messages, err := mq.client.Client.ZRangeByScore(mq.ctx, delayedQueueName, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return err
	}

	for _, msgData := range messages {
		// Move to main queue
		err := mq.client.Client.LPush(mq.ctx, queueName, msgData).Err()
		if err != nil {
			continue
		}

		// Remove from delayed queue
		mq.client.Client.ZRem(mq.ctx, delayedQueueName, msgData)
	}

	return nil
}

// GetQueueSize returns the size of a queue
func (mq *MessageQueue) GetQueueSize(queueName string) (int64, error) {
	return mq.client.Client.LLen(mq.ctx, queueName).Result()
}

// GetDelayedQueueSize returns the size of a delayed queue
func (mq *MessageQueue) GetDelayedQueueSize(queueName string) (int64, error) {
	delayedQueueName := fmt.Sprintf("%s:delayed", queueName)
	return mq.client.Client.ZCard(mq.ctx, delayedQueueName).Result()
}

// PubSub represents Redis pub/sub functionality
type PubSub struct {
	client      *RedisClient
	ctx         context.Context
	pubsub      *redis.PubSub
	subscribers map[string][]chan *redis.Message
	mu          sync.RWMutex
}

// NewPubSub creates a new pub/sub instance
func NewPubSub(ctx context.Context) (*PubSub, error) {
	client, err := NewRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	return &PubSub{
		client:      client,
		ctx:         ctx,
		subscribers: make(map[string][]chan *redis.Message),
	}, nil
}

// Publish publishes a message to a channel
func (ps *PubSub) Publish(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return ps.client.Client.Publish(ps.ctx, channel, data).Err()
}

// Subscribe subscribes to one or more channels (only available for standalone Redis)
func (ps *PubSub) Subscribe(channels ...string) (*redis.PubSub, error) {
	if client, ok := ps.client.Client.(*redis.Client); ok {
		return client.Subscribe(ps.ctx, channels...), nil
	}
	return nil, fmt.Errorf("subscribe not supported in cluster mode")
}

// PSubscribe subscribes to channels matching a pattern (only available for standalone Redis)
func (ps *PubSub) PSubscribe(patterns ...string) (*redis.PubSub, error) {
	if client, ok := ps.client.Client.(*redis.Client); ok {
		return client.PSubscribe(ps.ctx, patterns...), nil
	}
	return nil, fmt.Errorf("psubscribe not supported in cluster mode")
}

// MessageHandler represents a message handler function
type MessageHandler func(channel string, message []byte) error

// SubscribeWithHandler subscribes to channels with a handler function
func (ps *PubSub) SubscribeWithHandler(handler MessageHandler, channels ...string) error {
	pubsub, err := ps.Subscribe(channels...)
	if err != nil {
		return err
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		if err := handler(msg.Channel, []byte(msg.Payload)); err != nil {
			log.Printf("Error handling message on channel %s: %v", msg.Channel, err)
		}
	}

	return nil
}

// Worker represents a queue worker
type Worker struct {
	mq          *MessageQueue
	queueName   string
	concurrency int
	handlers    map[string]MessageHandler
	quit        chan struct{}
	wg          sync.WaitGroup
}

// NewWorker creates a new queue worker
func NewWorker(ctx context.Context, queueName string, concurrency int) (*Worker, error) {
	mq, err := NewMessageQueue(ctx)
	if err != nil {
		return nil, err
	}

	return &Worker{
		mq:          mq,
		queueName:   queueName,
		concurrency: concurrency,
		handlers:    make(map[string]MessageHandler),
		quit:        make(chan struct{}),
	}, nil
}

// RegisterHandler registers a handler for a message type
func (w *Worker) RegisterHandler(msgType string, handler MessageHandler) {
	w.handlers[msgType] = handler
}

// Start starts the worker
func (w *Worker) Start() {
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.processMessages()
	}

	// Start delayed message processor
	w.wg.Add(1)
	go w.processDelayedMessages()
}

// Stop stops the worker
func (w *Worker) Stop() {
	close(w.quit)
	w.wg.Wait()
}

// processMessages processes messages from the queue
func (w *Worker) processMessages() {
	defer w.wg.Done()

	for {
		select {
		case <-w.quit:
			return
		default:
			message, err := w.mq.Dequeue(w.queueName)
			if err != nil {
				if err == redis.Nil {
					continue // No message available
				}
				log.Printf("Error dequeuing message: %v", err)
				continue
			}

			if handler, exists := w.handlers[message.Type]; exists {
				payload, _ := json.Marshal(message.Payload)
				if err := handler(message.Type, payload); err != nil {
					log.Printf("Error processing message %s: %v", message.ID, err)

					// Requeue with delay
					opts := DefaultQueueOptions()
					w.mq.RequeueFailed(w.queueName, message, opts.RetryDelay)
				}
			} else {
				log.Printf("No handler found for message type: %s", message.Type)
			}
		}
	}
}

// processDelayedMessages processes delayed messages
func (w *Worker) processDelayedMessages() {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.quit:
			return
		case <-ticker.C:
			if err := w.mq.ProcessDelayedMessages(w.queueName); err != nil {
				log.Printf("Error processing delayed messages: %v", err)
			}
		}
	}
}

// Close closes the message queue
func (mq *MessageQueue) Close() error {
	return mq.client.Close()
}

// Close closes the pub/sub
func (ps *PubSub) Close() error {
	if ps.pubsub != nil {
		ps.pubsub.Close()
	}
	return ps.client.Close()
}

// Helper function to generate unique IDs
func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// Queue names constants
// const (
// 	TaskQueue         = "task_queue"
// 	EmailQueue        = "email_queue"
// 	NotificationQueue = "notification_queue"
// 	AnalyticsQueue    = "analytics_queue"
// )

// Message types constants
// const (
// 	TaskCreatedMsg      = "task.created"
// 	TaskUpdatedMsg      = "task.updated"
// 	TaskDeletedMsg      = "task.deleted"
// 	EmailSendMsg        = "email.send"
// 	NotificationSendMsg = "notification.send"
// 	AnalyticsEventMsg   = "analytics.event"
// )
