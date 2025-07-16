package models

import (
	"sync"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	MaxFileSize               = 10 << 20 // 10MB
	AllowedTypes              = ".pdf,.jpeg,.jpg,"
	StatusPending  TaskStatus = "pending"
	StatusReady    TaskStatus = "ready"
	StatusFailed   TaskStatus = "failed"
	StatusComplete TaskStatus = "Completed!"
	StatusWorking  TaskStatus = "working"
	StorageDir     string     = "./storage"
	BaseUrl        string     = "http://localhost:8080"
)

type Task struct {
	ID         uuid.UUID
	Links      []string
	Files      []FileInfo
	Status     TaskStatus
	ArchiveURL string
	Errors     []string
	Mu         sync.Mutex
}

type FileInfo struct {
	OriginalURL string
	FilePath    string
	Type        string
	Size        int64
	IsValid     bool
}
