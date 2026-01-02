package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestJobStatus_IsValid はIsValidメソッドが有効なステータスを正しく判定することをテストする
func TestJobStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		// 正常系: 有効なステータス
		{"pendingは有効", JobStatusPending, true},
		{"processingは有効", JobStatusProcessing, true},
		{"completedは有効", JobStatusCompleted, true},
		{"failedは有効", JobStatusFailed, true},
		// 異常系: 無効なステータス
		{"空文字列は無効", JobStatus(""), false},
		{"invalidは無効", JobStatus("invalid"), false},
		{"unknownは無効", JobStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestJobStatus_IsTerminal はIsTerminalメソッドが終端ステータスを正しく判定することをテストする
func TestJobStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"pendingは非終端", JobStatusPending, false},
		{"processingは非終端", JobStatusProcessing, false},
		{"completedは終端", JobStatusCompleted, true},
		{"failedは終端", JobStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsTerminal()
			if got != tt.want {
				t.Errorf("IsTerminal() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestNewClusterJob はNewClusterJobが正しい初期値でClusterJobを生成することをテストする
func TestNewClusterJob(t *testing.T) {
	priority := int32(10)

	before := time.Now()
	job := NewClusterJob(priority)
	after := time.Now()

	require.NotNil(t, job, "ClusterJobがnilです")

	// IDが設定されていることを確認
	if job.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("IDがゼロ値です")
	}

	if job.Status != JobStatusPending {
		t.Errorf("Status = %q, 期待値 %q", job.Status, JobStatusPending)
	}

	if job.Priority != priority {
		t.Errorf("Priority = %d, 期待値 %d", job.Priority, priority)
	}

	// CreatedAtが適切な時間範囲内であることを確認
	if job.CreatedAt.Before(before) || job.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, 期待範囲 %v - %v", job.CreatedAt, before, after)
	}

	if job.StartedAt != nil {
		t.Error("StartedAtがnilではありません")
	}

	if job.CompletedAt != nil {
		t.Error("CompletedAtがnilではありません")
	}

	if job.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q, 期待値 空文字列", job.ErrorMessage)
	}
}

// TestNewClusterJob_ZeroPriority は優先度0でも正常にジョブが作成されることをテストする
func TestNewClusterJob_ZeroPriority(t *testing.T) {
	job := NewClusterJob(0)

	require.NotNil(t, job, "ClusterJobがnilです")

	if job.Priority != 0 {
		t.Errorf("Priority = %d, 期待値 0", job.Priority)
	}
}

// TestNewClusterJob_NegativePriority は負の優先度でもジョブが作成されることをテストする
func TestNewClusterJob_NegativePriority(t *testing.T) {
	job := NewClusterJob(-5)

	require.NotNil(t, job, "ClusterJobがnilです")

	if job.Priority != -5 {
		t.Errorf("Priority = %d, 期待値 -5", job.Priority)
	}
}

// TestClusterJob_Start はStartメソッドがステータスをProcessingに変更しStartedAtを設定することをテストする
func TestClusterJob_Start(t *testing.T) {
	job := NewClusterJob(1)

	before := time.Now()
	job.Start()
	after := time.Now()

	if job.Status != JobStatusProcessing {
		t.Errorf("Status = %q, 期待値 %q", job.Status, JobStatusProcessing)
	}

	if job.StartedAt == nil {
		t.Error("StartedAtがnilです")
	}

	if job.StartedAt.Before(before) || job.StartedAt.After(after) {
		t.Errorf("StartedAt = %v, 期待範囲 %v - %v", *job.StartedAt, before, after)
	}
}

// TestClusterJob_Complete はCompleteメソッドがステータスをCompletedに変更しCompletedAtを設定することをテストする
func TestClusterJob_Complete(t *testing.T) {
	job := NewClusterJob(1)
	job.Start()

	before := time.Now()
	job.Complete()
	after := time.Now()

	if job.Status != JobStatusCompleted {
		t.Errorf("Status = %q, 期待値 %q", job.Status, JobStatusCompleted)
	}

	if job.CompletedAt == nil {
		t.Error("CompletedAtがnilです")
	}

	if job.CompletedAt.Before(before) || job.CompletedAt.After(after) {
		t.Errorf("CompletedAt = %v, 期待範囲 %v - %v", *job.CompletedAt, before, after)
	}
}

// TestClusterJob_Fail はFailメソッドがステータスをFailedに変更しエラーメッセージを設定することをテストする
func TestClusterJob_Fail(t *testing.T) {
	job := NewClusterJob(1)
	job.Start()
	errorMessage := "クラスター計算に失敗しました"

	before := time.Now()
	job.Fail(errorMessage)
	after := time.Now()

	if job.Status != JobStatusFailed {
		t.Errorf("Status = %q, 期待値 %q", job.Status, JobStatusFailed)
	}

	if job.ErrorMessage != errorMessage {
		t.Errorf("ErrorMessage = %q, 期待値 %q", job.ErrorMessage, errorMessage)
	}

	if job.CompletedAt == nil {
		t.Error("CompletedAtがnilです")
	}

	if job.CompletedAt.Before(before) || job.CompletedAt.After(after) {
		t.Errorf("CompletedAt = %v, 期待範囲 %v - %v", *job.CompletedAt, before, after)
	}
}

// TestClusterJob_Fail_EmptyMessage は空のエラーメッセージでも正常に動作することをテストする
func TestClusterJob_Fail_EmptyMessage(t *testing.T) {
	job := NewClusterJob(1)
	job.Start()

	job.Fail("")

	if job.Status != JobStatusFailed {
		t.Errorf("Status = %q, 期待値 %q", job.Status, JobStatusFailed)
	}

	if job.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q, 期待値 空文字列", job.ErrorMessage)
	}
}

// TestClusterJob_IsPending はIsPendingメソッドがPendingステータスのみtrueを返すことをテストする
func TestClusterJob_IsPending(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"pendingはtrue", JobStatusPending, true},
		{"processingはfalse", JobStatusProcessing, false},
		{"completedはfalse", JobStatusCompleted, false},
		{"failedはfalse", JobStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ClusterJob{Status: tt.status}
			if got := job.IsPending(); got != tt.want {
				t.Errorf("IsPending() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestClusterJob_IsProcessing はIsProcessingメソッドがProcessingステータスのみtrueを返すことをテストする
func TestClusterJob_IsProcessing(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"pendingはfalse", JobStatusPending, false},
		{"processingはtrue", JobStatusProcessing, true},
		{"completedはfalse", JobStatusCompleted, false},
		{"failedはfalse", JobStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ClusterJob{Status: tt.status}
			if got := job.IsProcessing(); got != tt.want {
				t.Errorf("IsProcessing() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestClusterJob_IsCompleted はIsCompletedメソッドがCompletedステータスのみtrueを返すことをテストする
func TestClusterJob_IsCompleted(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"pendingはfalse", JobStatusPending, false},
		{"processingはfalse", JobStatusProcessing, false},
		{"completedはtrue", JobStatusCompleted, true},
		{"failedはfalse", JobStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ClusterJob{Status: tt.status}
			if got := job.IsCompleted(); got != tt.want {
				t.Errorf("IsCompleted() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestClusterJob_IsFailed はIsFailedメソッドがFailedステータスのみtrueを返すことをテストする
func TestClusterJob_IsFailed(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{"pendingはfalse", JobStatusPending, false},
		{"processingはfalse", JobStatusProcessing, false},
		{"completedはfalse", JobStatusCompleted, false},
		{"failedはtrue", JobStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &ClusterJob{Status: tt.status}
			if got := job.IsFailed(); got != tt.want {
				t.Errorf("IsFailed() = %v, 期待値 %v", got, tt.want)
			}
		})
	}
}

// TestClusterJob_StateTransitions はジョブの状態遷移が正しく動作することをテストする
func TestClusterJob_StateTransitions(t *testing.T) {
	// Pending -> Processing -> Completed の遷移
	t.Run("正常フロー: Pending -> Processing -> Completed", func(t *testing.T) {
		job := NewClusterJob(1)

		if !job.IsPending() {
			t.Error("初期状態がPendingではありません")
		}

		job.Start()
		if !job.IsProcessing() {
			t.Error("Start後の状態がProcessingではありません")
		}

		job.Complete()
		if !job.IsCompleted() {
			t.Error("Complete後の状態がCompletedではありません")
		}
	})

	// Pending -> Processing -> Failed の遷移
	t.Run("失敗フロー: Pending -> Processing -> Failed", func(t *testing.T) {
		job := NewClusterJob(1)

		job.Start()
		if !job.IsProcessing() {
			t.Error("Start後の状態がProcessingではありません")
		}

		job.Fail("エラーが発生しました")
		if !job.IsFailed() {
			t.Error("Fail後の状態がFailedではありません")
		}
	})
}
