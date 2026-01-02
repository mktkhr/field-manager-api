// Package entity はクラスタリング機能のドメインエンティティを定義する
package entity

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus はクラスタージョブのステータス
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// IsValid はステータスが有効かどうかを判定する
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusPending, JobStatusProcessing, JobStatusCompleted, JobStatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal はステータスが終了状態かどうかを判定する
func (s JobStatus) IsTerminal() bool {
	return s == JobStatusCompleted || s == JobStatusFailed
}

// ClusterJob はクラスタリングジョブのエンティティ
type ClusterJob struct {
	ID              uuid.UUID
	Status          JobStatus
	Priority        int32
	AffectedH3Cells []string // 影響を受けたH3セル(nil=全範囲再計算)
	CreatedAt       time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	ErrorMessage    string
}

// NewClusterJob は新しいClusterJobを作成する(全範囲再計算)
func NewClusterJob(priority int32) *ClusterJob {
	now := time.Now()
	return &ClusterJob{
		ID:        uuid.New(),
		Status:    JobStatusPending,
		Priority:  priority,
		CreatedAt: now,
	}
}

// NewClusterJobWithAffectedCells は影響セル指定でClusterJobを作成する
func NewClusterJobWithAffectedCells(priority int32, affectedCells []string) *ClusterJob {
	now := time.Now()
	return &ClusterJob{
		ID:              uuid.New(),
		Status:          JobStatusPending,
		Priority:        priority,
		AffectedH3Cells: affectedCells,
		CreatedAt:       now,
	}
}

// IsFullRecalculation は全範囲再計算かどうかを判定する
func (j *ClusterJob) IsFullRecalculation() bool {
	return len(j.AffectedH3Cells) == 0
}

// Start はジョブを処理中状態に遷移する
func (j *ClusterJob) Start() {
	now := time.Now()
	j.Status = JobStatusProcessing
	j.StartedAt = &now
}

// Complete はジョブを完了状態に遷移する
func (j *ClusterJob) Complete() {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
}

// Fail はジョブを失敗状態に遷移する
func (j *ClusterJob) Fail(errorMessage string) {
	now := time.Now()
	j.Status = JobStatusFailed
	j.CompletedAt = &now
	j.ErrorMessage = errorMessage
}

// IsPending はジョブが保留中かどうかを判定する
func (j *ClusterJob) IsPending() bool {
	return j.Status == JobStatusPending
}

// IsProcessing はジョブが処理中かどうかを判定する
func (j *ClusterJob) IsProcessing() bool {
	return j.Status == JobStatusProcessing
}

// IsCompleted はジョブが完了しているかどうかを判定する
func (j *ClusterJob) IsCompleted() bool {
	return j.Status == JobStatusCompleted
}

// IsFailed はジョブが失敗しているかどうかを判定する
func (j *ClusterJob) IsFailed() bool {
	return j.Status == JobStatusFailed
}
