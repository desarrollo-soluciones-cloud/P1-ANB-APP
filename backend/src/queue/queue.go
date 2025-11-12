package queue

import (
	"context"
	"time"
)

// TaskPayload representa el payload de una tarea
type TaskPayload struct {
	VideoID uint `json:"video_id"`
}

// Task representa una tarea en la cola
type Task struct {
	ID      string
	Type    string
	Payload TaskPayload
}

// QueueClient interfaz para abstraer el sistema de colas
type QueueClient interface {
	// EnqueueTask encola una nueva tarea
	EnqueueTask(ctx context.Context, taskType string, payload TaskPayload, maxRetry int, timeout time.Duration) (string, error)
	// Close cierra la conexión con el sistema de colas
	Close() error
}

// QueueConsumer interfaz para consumir tareas de la cola
type QueueConsumer interface {
	// ReceiveTask recibe la siguiente tarea disponible
	ReceiveTask(ctx context.Context) (*Task, error)
	// CompleteTask marca una tarea como completada
	CompleteTask(ctx context.Context, task *Task) error
	// FailTask marca una tarea como fallida
	FailTask(ctx context.Context, task *Task) error
	// Close cierra la conexión con el sistema de colas
	Close() error
}
