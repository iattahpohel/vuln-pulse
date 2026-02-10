package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/atta/vulnpulse/internal/service"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/atta/vulnpulse/pkg/queue"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Worker processes jobs from the queue
type Worker struct {
	queue        *queue.Client
	matchService *service.MatchService
	log          *logger.Logger
	queueName    string
}

// NewWorker creates a new worker instance
func NewWorker(
	queueClient *queue.Client,
	matchService *service.MatchService,
	log *logger.Logger,
	queueName string,
) *Worker {
	return &Worker{
		queue:        queueClient,
		matchService: matchService,
		log:          log,
		queueName:    queueName,
	}
}

// Start begins processing jobs
func (w *Worker) Start(ctx context.Context) error {
	w.log.Info("starting worker", "queue", w.queueName)

	msgs, err := w.queue.Consume(w.queueName)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	// Process messages
	for msg := range msgs {
		w.processMessage(ctx, msg)
	}

	return nil
}

func (w *Worker) processMessage(ctx context.Context, delivery amqp.Delivery) {
	w.log.Info("processing message", "delivery_tag", delivery.DeliveryTag)

	var msg queue.Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		w.log.Error("failed to unmarshal message", "error", err)
		delivery.Nack(false, false) // Don't requeue malformed messages
		return
	}

	var err error
	switch msg.Type {
	case "vuln.ingested":
		err = w.handleVulnIngested(ctx, msg)
	case "asset.changed":
		err = w.handleAssetChanged(ctx, msg)
	case "webhook.dispatch":
		// Webhook dispatcher would handle this
		w.log.Info("webhook dispatch event received (skipping in worker)")
		delivery.Ack(false)
		return
	default:
		w.log.Warn("unknown message type", "type", msg.Type)
		delivery.Nack(false, false)
		return
	}

	if err != nil {
		w.log.Error("failed to process message", "type", msg.Type, "error", err)
		// Requeue for retry
		delivery.Nack(false, true)
		return
	}

	delivery.Ack(false)
	w.log.Info("message processed successfully", "type", msg.Type)
}

func (w *Worker) handleVulnIngested(ctx context.Context, msg queue.Message) error {
	vulnIDStr, ok := msg.Payload["vulnerability_id"].(string)
	if !ok {
		return fmt.Errorf("invalid vulnerability_id in payload")
	}

	vulnID, err := uuid.Parse(vulnIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse vulnerability_id: %w", err)
	}

	return w.matchService.MatchVulnerability(ctx, vulnID)
}

func (w *Worker) handleAssetChanged(ctx context.Context, msg queue.Message) error {
	assetIDStr, ok := msg.Payload["asset_id"].(string)
	if !ok {
		return fmt.Errorf("invalid asset_id in payload")
	}

	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse asset_id: %w", err)
	}

	return w.matchService.MatchAsset(ctx, assetID)
}
