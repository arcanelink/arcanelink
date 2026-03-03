package cleanup

import (
	"time"

	"github.com/arcane/arcanelink/internal/presence/repository"
	"github.com/arcane/arcanelink/pkg/logger"
	"go.uber.org/zap"
)

type CleanupJob struct {
	repo              *repository.PresenceRepository
	inactiveTimeout   time.Duration
	cleanupInterval   time.Duration
	deleteOlderThan   time.Duration
	stopChan          chan struct{}
}

func NewCleanupJob(repo *repository.PresenceRepository) *CleanupJob {
	return &CleanupJob{
		repo:            repo,
		inactiveTimeout: 60 * time.Second,  // Mark offline after 60 seconds
		cleanupInterval: 30 * time.Second,  // Run cleanup every 30 seconds
		deleteOlderThan: 7 * 24 * time.Hour, // Delete records older than 7 days
		stopChan:        make(chan struct{}),
	}
}

// Start starts the cleanup job
func (j *CleanupJob) Start() {
	logger.Info("Starting presence cleanup job",
		zap.Duration("inactive_timeout", j.inactiveTimeout),
		zap.Duration("cleanup_interval", j.cleanupInterval))

	ticker := time.NewTicker(j.cleanupInterval)
	defer ticker.Stop()

	// Run immediately on start
	j.runCleanup()

	for {
		select {
		case <-ticker.C:
			j.runCleanup()
		case <-j.stopChan:
			logger.Info("Stopping presence cleanup job")
			return
		}
	}
}

// Stop stops the cleanup job
func (j *CleanupJob) Stop() {
	close(j.stopChan)
}

// runCleanup performs the cleanup operations
func (j *CleanupJob) runCleanup() {
	// Mark inactive users as offline
	offlineCount, err := j.repo.MarkOfflineIfInactive(j.inactiveTimeout)
	if err != nil {
		logger.Error("Failed to mark inactive users offline", zap.Error(err))
	} else if offlineCount > 0 {
		logger.Debug("Marked users as offline", zap.Int("count", offlineCount))
	}

	// Delete old presence records (run less frequently)
	if time.Now().Unix()%300 == 0 { // Every 5 minutes
		deletedCount, err := j.repo.DeleteOldPresence(j.deleteOlderThan)
		if err != nil {
			logger.Error("Failed to delete old presence records", zap.Error(err))
		} else if deletedCount > 0 {
			logger.Info("Deleted old presence records", zap.Int("count", deletedCount))
		}
	}

	// Log online count periodically
	if time.Now().Unix()%60 == 0 { // Every minute
		onlineCount, err := j.repo.GetOnlineCount()
		if err != nil {
			logger.Error("Failed to get online count", zap.Error(err))
		} else {
			logger.Info("Current online users", zap.Int("count", onlineCount))
		}
	}
}
