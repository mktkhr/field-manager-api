package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ImportStatus はインポートジョブのステータスを表す
type ImportStatus string

const (
	ImportStatusPending            ImportStatus = "pending"
	ImportStatusProcessing         ImportStatus = "processing"
	ImportStatusCompleted          ImportStatus = "completed"
	ImportStatusFailed             ImportStatus = "failed"
	ImportStatusPartiallyCompleted ImportStatus = "partially_completed"
)

// IsValid はステータスが有効かどうかを判定する
func (s ImportStatus) IsValid() bool {
	switch s {
	case ImportStatusPending, ImportStatusProcessing, ImportStatusCompleted, ImportStatusFailed, ImportStatusPartiallyCompleted:
		return true
	}
	return false
}

// String はステータスを文字列として返す
func (s ImportStatus) String() string {
	return string(s)
}

// ImportJob はインポートジョブエンティティ
type ImportJob struct {
	ID                 uuid.UUID
	CityCode           string
	Status             ImportStatus
	TotalRecords       *int32
	ProcessedRecords   int32
	FailedRecords      int32
	LastProcessedBatch int32
	S3Key              *string
	ExecutionArn       *string
	ErrorMessage       *string
	FailedRecordIDs    []string
	CreatedAt          time.Time
	StartedAt          *time.Time
	CompletedAt        *time.Time
}

// NewImportJob は新しいインポートジョブを作成する
func NewImportJob(cityCode string) *ImportJob {
	return &ImportJob{
		ID:               uuid.New(),
		CityCode:         cityCode,
		Status:           ImportStatusPending,
		ProcessedRecords: 0,
		FailedRecords:    0,
		CreatedAt:        time.Now(),
	}
}

// CanTransitionTo は指定のステータスに遷移可能かどうかを判定する
func (j *ImportJob) CanTransitionTo(newStatus ImportStatus) bool {
	switch j.Status {
	case ImportStatusPending:
		return newStatus == ImportStatusProcessing || newStatus == ImportStatusFailed
	case ImportStatusProcessing:
		return newStatus == ImportStatusCompleted || newStatus == ImportStatusFailed || newStatus == ImportStatusPartiallyCompleted
	case ImportStatusCompleted, ImportStatusFailed, ImportStatusPartiallyCompleted:
		return false
	}
	return false
}

// TransitionTo はステータスを遷移させる
func (j *ImportJob) TransitionTo(newStatus ImportStatus) error {
	if !j.CanTransitionTo(newStatus) {
		return ErrInvalidStatusTransition
	}

	now := time.Now()
	j.Status = newStatus

	switch newStatus {
	case ImportStatusProcessing:
		j.StartedAt = &now
	case ImportStatusCompleted, ImportStatusFailed, ImportStatusPartiallyCompleted:
		j.CompletedAt = &now
	}

	return nil
}

// SetS3Key はS3キーを設定する
func (j *ImportJob) SetS3Key(s3Key string) {
	j.S3Key = &s3Key
}

// SetExecutionArn は実行ARNを設定する
func (j *ImportJob) SetExecutionArn(arn string) {
	j.ExecutionArn = &arn
}

// SetTotalRecords は総レコード数を設定する
func (j *ImportJob) SetTotalRecords(total int32) {
	j.TotalRecords = &total
}

// UpdateProgress は進捗を更新する
func (j *ImportJob) UpdateProgress(processed, failed, batch int32) {
	j.ProcessedRecords = processed
	j.FailedRecords = failed
	j.LastProcessedBatch = batch
}

// SetError はエラー情報を設定する
func (j *ImportJob) SetError(message string, failedIDs []string) {
	j.ErrorMessage = &message
	j.FailedRecordIDs = failedIDs
	now := time.Now()
	j.CompletedAt = &now
	j.Status = ImportStatusFailed
}

// IsTerminal はジョブが終了状態かどうかを判定する
func (j *ImportJob) IsTerminal() bool {
	return j.Status == ImportStatusCompleted || j.Status == ImportStatusFailed || j.Status == ImportStatusPartiallyCompleted
}

// IsRunning はジョブが実行中かどうかを判定する
func (j *ImportJob) IsRunning() bool {
	return j.Status == ImportStatusProcessing
}

// Start はジョブを開始する
func (j *ImportJob) Start() error {
	return j.TransitionTo(ImportStatusProcessing)
}

// Complete はジョブを完了する
func (j *ImportJob) Complete() error {
	return j.TransitionTo(ImportStatusCompleted)
}

// Fail はジョブを失敗状態にする
func (j *ImportJob) Fail(message string) error {
	j.ErrorMessage = &message
	return j.TransitionTo(ImportStatusFailed)
}

// PartialComplete はジョブを部分完了状態にする
func (j *ImportJob) PartialComplete() error {
	return j.TransitionTo(ImportStatusPartiallyCompleted)
}

// Duration はジョブの実行時間を返す
func (j *ImportJob) Duration() *time.Duration {
	if j.StartedAt == nil || j.CompletedAt == nil {
		return nil
	}
	d := j.CompletedAt.Sub(*j.StartedAt)
	return &d
}

// Progress は進捗率を返す(0-100)
// 処理済み + 失敗を合計した完了分で計算する
func (j *ImportJob) Progress() float64 {
	if j.TotalRecords == nil || *j.TotalRecords == 0 {
		return 0
	}
	completed := j.ProcessedRecords + j.FailedRecords
	return float64(completed) / float64(*j.TotalRecords) * 100
}

// FailedRecordIDsJSON は失敗したレコードIDをJSON形式で返す
func (j *ImportJob) FailedRecordIDsJSON() (json.RawMessage, error) {
	if len(j.FailedRecordIDs) == 0 {
		return nil, nil
	}
	return json.Marshal(j.FailedRecordIDs)
}
