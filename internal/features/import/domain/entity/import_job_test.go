package entity

import (
	"testing"
	"time"
)

func TestNewImportJob(t *testing.T) {
	cityCode := "163210"

	job := NewImportJob(cityCode)

	if job.CityCode != cityCode {
		t.Errorf("CityCode = %q, want %q", job.CityCode, cityCode)
	}
	if job.Status != ImportStatusPending {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusPending)
	}
	if job.ProcessedRecords != 0 {
		t.Errorf("ProcessedRecords = %d, want 0", job.ProcessedRecords)
	}
	if job.FailedRecords != 0 {
		t.Errorf("FailedRecords = %d, want 0", job.FailedRecords)
	}
	if job.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestImportJobStart(t *testing.T) {
	job := NewImportJob("163210")

	job.Start()

	if job.Status != ImportStatusProcessing {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusProcessing)
	}
	if job.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}
}

func TestImportJobComplete(t *testing.T) {
	job := NewImportJob("163210")
	job.Start()

	job.Complete()

	if job.Status != ImportStatusCompleted {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusCompleted)
	}
	if job.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestImportJobFail(t *testing.T) {
	job := NewImportJob("163210")
	job.Start()
	errorMessage := "Something went wrong"

	job.Fail(errorMessage)

	if job.Status != ImportStatusFailed {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusFailed)
	}
	if job.ErrorMessage == nil {
		t.Error("ErrorMessage should not be nil")
	}
	if *job.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %q, want %q", *job.ErrorMessage, errorMessage)
	}
	if job.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestImportJobPartialComplete(t *testing.T) {
	job := NewImportJob("163210")
	job.Start()

	job.PartialComplete()

	if job.Status != ImportStatusPartiallyCompleted {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusPartiallyCompleted)
	}
	if job.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestImportJobProgress(t *testing.T) {
	tests := []struct {
		name             string
		totalRecords     *int32
		processedRecords int32
		failedRecords    int32
		want             float64
	}{
		{
			name:             "no total records",
			totalRecords:     nil,
			processedRecords: 0,
			failedRecords:    0,
			want:             0,
		},
		{
			name:             "zero total records",
			totalRecords:     int32Ptr(0),
			processedRecords: 0,
			failedRecords:    0,
			want:             0,
		},
		{
			name:             "50% progress",
			totalRecords:     int32Ptr(100),
			processedRecords: 50,
			failedRecords:    0,
			want:             50,
		},
		{
			name:             "100% progress with processed only",
			totalRecords:     int32Ptr(100),
			processedRecords: 100,
			failedRecords:    0,
			want:             100,
		},
		{
			name:             "100% progress with processed and failed",
			totalRecords:     int32Ptr(100),
			processedRecords: 80,
			failedRecords:    20,
			want:             100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{
				TotalRecords:     tt.totalRecords,
				ProcessedRecords: tt.processedRecords,
				FailedRecords:    tt.failedRecords,
			}

			got := job.Progress()
			if got != tt.want {
				t.Errorf("Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportJobIsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status ImportStatus
		want   bool
	}{
		{
			name:   "pending is not terminal",
			status: ImportStatusPending,
			want:   false,
		},
		{
			name:   "processing is not terminal",
			status: ImportStatusProcessing,
			want:   false,
		},
		{
			name:   "completed is terminal",
			status: ImportStatusCompleted,
			want:   true,
		},
		{
			name:   "failed is terminal",
			status: ImportStatusFailed,
			want:   true,
		},
		{
			name:   "partially_completed is terminal",
			status: ImportStatusPartiallyCompleted,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{Status: tt.status}
			if got := job.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportJobIsRunning(t *testing.T) {
	tests := []struct {
		name   string
		status ImportStatus
		want   bool
	}{
		{
			name:   "pending is not running",
			status: ImportStatusPending,
			want:   false,
		},
		{
			name:   "processing is running",
			status: ImportStatusProcessing,
			want:   true,
		},
		{
			name:   "completed is not running",
			status: ImportStatusCompleted,
			want:   false,
		},
		{
			name:   "failed is not running",
			status: ImportStatusFailed,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{Status: tt.status}
			if got := job.IsRunning(); got != tt.want {
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportJobDuration(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	tests := []struct {
		name        string
		startedAt   *time.Time
		completedAt *time.Time
		wantNil     bool
	}{
		{
			name:        "not started",
			startedAt:   nil,
			completedAt: nil,
			wantNil:     true,
		},
		{
			name:        "started but not completed",
			startedAt:   &oneHourAgo,
			completedAt: nil,
			wantNil:     true,
		},
		{
			name:        "completed",
			startedAt:   &oneHourAgo,
			completedAt: &now,
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{
				StartedAt:   tt.startedAt,
				CompletedAt: tt.completedAt,
			}

			got := job.Duration()
			if tt.wantNil {
				if got != nil {
					t.Errorf("Duration() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("Duration() = nil, want non-nil")
				} else if *got < time.Hour-time.Second || *got > time.Hour+time.Second {
					t.Errorf("Duration() = %v, want approximately 1 hour", *got)
				}
			}
		})
	}
}

func int32Ptr(v int32) *int32 {
	return &v
}

func TestImportStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ImportStatus
		want   bool
	}{
		{"pending is valid", ImportStatusPending, true},
		{"processing is valid", ImportStatusProcessing, true},
		{"completed is valid", ImportStatusCompleted, true},
		{"failed is valid", ImportStatusFailed, true},
		{"partially_completed is valid", ImportStatusPartiallyCompleted, true},
		{"invalid status", ImportStatus("invalid"), false},
		{"empty status", ImportStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status ImportStatus
		want   string
	}{
		{"pending", ImportStatusPending, "pending"},
		{"processing", ImportStatusProcessing, "processing"},
		{"completed", ImportStatusCompleted, "completed"},
		{"failed", ImportStatusFailed, "failed"},
		{"partially_completed", ImportStatusPartiallyCompleted, "partially_completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportJob_SetS3Key(t *testing.T) {
	job := NewImportJob("163210")
	s3Key := "imports/163210/2024-01-01.json"

	job.SetS3Key(s3Key)

	if job.S3Key == nil {
		t.Error("S3Key should not be nil")
	}
	if *job.S3Key != s3Key {
		t.Errorf("S3Key = %q, want %q", *job.S3Key, s3Key)
	}
}

func TestImportJob_SetExecutionArn(t *testing.T) {
	job := NewImportJob("163210")
	arn := "arn:aws:states:ap-northeast-1:123456789012:execution:test:abc123"

	job.SetExecutionArn(arn)

	if job.ExecutionArn == nil {
		t.Error("ExecutionArn should not be nil")
	}
	if *job.ExecutionArn != arn {
		t.Errorf("ExecutionArn = %q, want %q", *job.ExecutionArn, arn)
	}
}

func TestImportJob_SetTotalRecords(t *testing.T) {
	job := NewImportJob("163210")
	total := int32(1000)

	job.SetTotalRecords(total)

	if job.TotalRecords == nil {
		t.Error("TotalRecords should not be nil")
	}
	if *job.TotalRecords != total {
		t.Errorf("TotalRecords = %d, want %d", *job.TotalRecords, total)
	}
}

func TestImportJob_UpdateProgress(t *testing.T) {
	job := NewImportJob("163210")

	job.UpdateProgress(50, 10, 5)

	if job.ProcessedRecords != 50 {
		t.Errorf("ProcessedRecords = %d, want %d", job.ProcessedRecords, 50)
	}
	if job.FailedRecords != 10 {
		t.Errorf("FailedRecords = %d, want %d", job.FailedRecords, 10)
	}
	if job.LastProcessedBatch != 5 {
		t.Errorf("LastProcessedBatch = %d, want %d", job.LastProcessedBatch, 5)
	}
}

func TestImportJob_SetError(t *testing.T) {
	job := NewImportJob("163210")
	job.Start()
	errorMessage := "Something went wrong"
	failedIDs := []string{"id1", "id2", "id3"}

	job.SetError(errorMessage, failedIDs)

	if job.ErrorMessage == nil {
		t.Error("ErrorMessage should not be nil")
	}
	if *job.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %q, want %q", *job.ErrorMessage, errorMessage)
	}
	if len(job.FailedRecordIDs) != 3 {
		t.Errorf("len(FailedRecordIDs) = %d, want %d", len(job.FailedRecordIDs), 3)
	}
	if job.Status != ImportStatusFailed {
		t.Errorf("Status = %q, want %q", job.Status, ImportStatusFailed)
	}
	if job.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestImportJob_FailedRecordIDsJSON(t *testing.T) {
	tests := []struct {
		name      string
		failedIDs []string
		wantNil   bool
		wantJSON  string
	}{
		{
			name:      "empty failed IDs",
			failedIDs: nil,
			wantNil:   true,
		},
		{
			name:      "empty slice",
			failedIDs: []string{},
			wantNil:   true,
		},
		{
			name:      "with failed IDs",
			failedIDs: []string{"id1", "id2", "id3"},
			wantNil:   false,
			wantJSON:  `["id1","id2","id3"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{FailedRecordIDs: tt.failedIDs}

			got, err := job.FailedRecordIDsJSON()
			if err != nil {
				t.Errorf("FailedRecordIDsJSON() error = %v", err)
				return
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("FailedRecordIDsJSON() = %s, want nil", got)
				}
			} else {
				if string(got) != tt.wantJSON {
					t.Errorf("FailedRecordIDsJSON() = %s, want %s", got, tt.wantJSON)
				}
			}
		})
	}
}

func TestImportJob_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name      string
		current   ImportStatus
		target    ImportStatus
		canChange bool
	}{
		// From Pending
		{"pending to processing", ImportStatusPending, ImportStatusProcessing, true},
		{"pending to failed", ImportStatusPending, ImportStatusFailed, true},
		{"pending to completed", ImportStatusPending, ImportStatusCompleted, false},
		{"pending to partially_completed", ImportStatusPending, ImportStatusPartiallyCompleted, false},
		// From Processing
		{"processing to completed", ImportStatusProcessing, ImportStatusCompleted, true},
		{"processing to failed", ImportStatusProcessing, ImportStatusFailed, true},
		{"processing to partially_completed", ImportStatusProcessing, ImportStatusPartiallyCompleted, true},
		{"processing to pending", ImportStatusProcessing, ImportStatusPending, false},
		// From Completed
		{"completed to any", ImportStatusCompleted, ImportStatusPending, false},
		{"completed to processing", ImportStatusCompleted, ImportStatusProcessing, false},
		// From Failed
		{"failed to any", ImportStatusFailed, ImportStatusPending, false},
		{"failed to processing", ImportStatusFailed, ImportStatusProcessing, false},
		// From PartiallyCompleted
		{"partially_completed to any", ImportStatusPartiallyCompleted, ImportStatusPending, false},
		{"partially_completed to processing", ImportStatusPartiallyCompleted, ImportStatusProcessing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{Status: tt.current}
			if got := job.CanTransitionTo(tt.target); got != tt.canChange {
				t.Errorf("CanTransitionTo(%s) = %v, want %v", tt.target, got, tt.canChange)
			}
		})
	}
}

func TestImportJob_TransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current ImportStatus
		target  ImportStatus
		wantErr bool
	}{
		{"valid transition pending to processing", ImportStatusPending, ImportStatusProcessing, false},
		{"valid transition processing to completed", ImportStatusProcessing, ImportStatusCompleted, false},
		{"invalid transition pending to completed", ImportStatusPending, ImportStatusCompleted, true},
		{"invalid transition completed to processing", ImportStatusCompleted, ImportStatusProcessing, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ImportJob{Status: tt.current}
			err := job.TransitionTo(tt.target)

			if tt.wantErr {
				if err == nil {
					t.Error("TransitionTo() expected error, got nil")
				}
				if job.Status != tt.current {
					t.Errorf("Status changed to %s, should remain %s on error", job.Status, tt.current)
				}
			} else {
				if err != nil {
					t.Errorf("TransitionTo() error = %v", err)
				}
				if job.Status != tt.target {
					t.Errorf("Status = %s, want %s", job.Status, tt.target)
				}
			}
		})
	}
}

func TestImportJob_TransitionTo_SetsTimestamps(t *testing.T) {
	// Test StartedAt is set when transitioning to Processing
	job := NewImportJob("163210")
	err := job.TransitionTo(ImportStatusProcessing)
	if err != nil {
		t.Errorf("TransitionTo() error = %v", err)
	}
	if job.StartedAt == nil {
		t.Error("StartedAt should be set when transitioning to Processing")
	}

	// Test CompletedAt is set when transitioning to terminal states
	terminalStates := []ImportStatus{ImportStatusCompleted, ImportStatusFailed, ImportStatusPartiallyCompleted}
	for _, state := range terminalStates {
		job := NewImportJob("163210")
		job.Status = ImportStatusProcessing // Set to processing first
		err := job.TransitionTo(state)
		if err != nil {
			t.Errorf("TransitionTo(%s) error = %v", state, err)
		}
		if job.CompletedAt == nil {
			t.Errorf("CompletedAt should be set when transitioning to %s", state)
		}
	}
}
