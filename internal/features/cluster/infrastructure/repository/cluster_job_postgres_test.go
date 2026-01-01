package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mktkhr/field-manager-api/internal/features/cluster/domain/entity"
	"github.com/mktkhr/field-manager-api/internal/generated/sqlc"
	"github.com/stretchr/testify/require"
)

// TestConvertToClusterJobEntity はconvertToClusterJobEntityがsqlc.ClusterJobをエンティティに正しく変換することをテストする
func TestConvertToClusterJobEntity(t *testing.T) {
	now := time.Now()
	jobID := uuid.New()

	tests := []struct {
		name string
		job  *sqlc.ClusterJob
		want *entity.ClusterJob
	}{
		// 正常系: 必須フィールドのみ
		{
			name: "必須フィールドのみのジョブ",
			job: &sqlc.ClusterJob{
				ID:        jobID,
				Status:    "pending",
				Priority:  10,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.ClusterJob{
				ID:        jobID,
				Status:    entity.JobStatusPending,
				Priority:  10,
				CreatedAt: now,
			},
		},
		// 正常系: 処理中ジョブ
		{
			name: "処理中ジョブ",
			job: &sqlc.ClusterJob{
				ID:        jobID,
				Status:    "processing",
				Priority:  5,
				CreatedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
				StartedAt: pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.ClusterJob{
				ID:        jobID,
				Status:    entity.JobStatusProcessing,
				Priority:  5,
				CreatedAt: now.Add(-time.Hour),
			},
		},
		// 正常系: 完了済みジョブ
		{
			name: "完了済みジョブ",
			job: &sqlc.ClusterJob{
				ID:          jobID,
				Status:      "completed",
				Priority:    1,
				CreatedAt:   pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
				StartedAt:   pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
				CompletedAt: pgtype.Timestamptz{Time: now, Valid: true},
			},
			want: &entity.ClusterJob{
				ID:        jobID,
				Status:    entity.JobStatusCompleted,
				Priority:  1,
				CreatedAt: now.Add(-2 * time.Hour),
			},
		},
		// 正常系: 失敗ジョブ
		{
			name: "失敗ジョブ",
			job: &sqlc.ClusterJob{
				ID:           jobID,
				Status:       "failed",
				Priority:     0,
				CreatedAt:    pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
				StartedAt:    pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
				CompletedAt:  pgtype.Timestamptz{Time: now, Valid: true},
				ErrorMessage: stringPtr("エラーが発生しました"),
			},
			want: &entity.ClusterJob{
				ID:           jobID,
				Status:       entity.JobStatusFailed,
				Priority:     0,
				CreatedAt:    now.Add(-2 * time.Hour),
				ErrorMessage: "エラーが発生しました",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToClusterJobEntity(tt.job)

			require.NotNil(t, got, "結果がnilです")

			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, 期待値 %v", got.ID, tt.want.ID)
			}

			if got.Status != tt.want.Status {
				t.Errorf("Status = %v, 期待値 %v", got.Status, tt.want.Status)
			}

			if got.Priority != tt.want.Priority {
				t.Errorf("Priority = %d, 期待値 %d", got.Priority, tt.want.Priority)
			}

			// CreatedAtの比較(秒単位)
			if got.CreatedAt.Unix() != tt.want.CreatedAt.Unix() {
				t.Errorf("CreatedAt = %v, 期待値 %v", got.CreatedAt, tt.want.CreatedAt)
			}

			if got.ErrorMessage != tt.want.ErrorMessage {
				t.Errorf("ErrorMessage = %q, 期待値 %q", got.ErrorMessage, tt.want.ErrorMessage)
			}
		})
	}
}

// TestConvertToClusterJobEntity_StartedAt はStartedAtが正しく変換されることをテストする
func TestConvertToClusterJobEntity_StartedAt(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		startedAt    pgtype.Timestamptz
		wantNil      bool
		wantTimeUnix int64
	}{
		{
			name:      "StartedAtがnull",
			startedAt: pgtype.Timestamptz{Valid: false},
			wantNil:   true,
		},
		{
			name:         "StartedAtが設定済み",
			startedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			wantNil:      false,
			wantTimeUnix: now.Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &sqlc.ClusterJob{
				ID:        uuid.New(),
				Status:    "processing",
				Priority:  1,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt: tt.startedAt,
			}

			got := convertToClusterJobEntity(job)

			if tt.wantNil {
				if got.StartedAt != nil {
					t.Error("StartedAtがnilではありません")
				}
			} else {
				if got.StartedAt == nil {
					t.Error("StartedAtがnilです")
				} else if got.StartedAt.Unix() != tt.wantTimeUnix {
					t.Errorf("StartedAt = %v, 期待値 %v", got.StartedAt.Unix(), tt.wantTimeUnix)
				}
			}
		})
	}
}

// TestConvertToClusterJobEntity_CompletedAt はCompletedAtが正しく変換されることをテストする
func TestConvertToClusterJobEntity_CompletedAt(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		completedAt  pgtype.Timestamptz
		wantNil      bool
		wantTimeUnix int64
	}{
		{
			name:        "CompletedAtがnull",
			completedAt: pgtype.Timestamptz{Valid: false},
			wantNil:     true,
		},
		{
			name:         "CompletedAtが設定済み",
			completedAt:  pgtype.Timestamptz{Time: now, Valid: true},
			wantNil:      false,
			wantTimeUnix: now.Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &sqlc.ClusterJob{
				ID:          uuid.New(),
				Status:      "completed",
				Priority:    1,
				CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				StartedAt:   pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
				CompletedAt: tt.completedAt,
			}

			got := convertToClusterJobEntity(job)

			if tt.wantNil {
				if got.CompletedAt != nil {
					t.Error("CompletedAtがnilではありません")
				}
			} else {
				if got.CompletedAt == nil {
					t.Error("CompletedAtがnilです")
				} else if got.CompletedAt.Unix() != tt.wantTimeUnix {
					t.Errorf("CompletedAt = %v, 期待値 %v", got.CompletedAt.Unix(), tt.wantTimeUnix)
				}
			}
		})
	}
}

// TestConvertToClusterJobEntity_ErrorMessage はErrorMessageが正しく変換されることをテストする
func TestConvertToClusterJobEntity_ErrorMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		errorMessage *string
		wantMessage  string
	}{
		{
			name:         "ErrorMessageがnull",
			errorMessage: nil,
			wantMessage:  "",
		},
		{
			name:         "ErrorMessageが空文字列",
			errorMessage: stringPtr(""),
			wantMessage:  "",
		},
		{
			name:         "ErrorMessageが設定済み",
			errorMessage: stringPtr("クラスター計算に失敗しました"),
			wantMessage:  "クラスター計算に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &sqlc.ClusterJob{
				ID:           uuid.New(),
				Status:       "failed",
				Priority:     1,
				CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
				ErrorMessage: tt.errorMessage,
			}

			got := convertToClusterJobEntity(job)

			if got.ErrorMessage != tt.wantMessage {
				t.Errorf("ErrorMessage = %q, 期待値 %q", got.ErrorMessage, tt.wantMessage)
			}
		})
	}
}

// TestConvertToClusterJobEntity_AllStatuses は全てのステータスが正しく変換されることをテストする
func TestConvertToClusterJobEntity_AllStatuses(t *testing.T) {
	now := time.Now()

	statuses := []struct {
		input  string
		output entity.JobStatus
	}{
		{"pending", entity.JobStatusPending},
		{"processing", entity.JobStatusProcessing},
		{"completed", entity.JobStatusCompleted},
		{"failed", entity.JobStatusFailed},
	}

	for _, s := range statuses {
		t.Run(s.input, func(t *testing.T) {
			job := &sqlc.ClusterJob{
				ID:        uuid.New(),
				Status:    s.input,
				Priority:  1,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}

			got := convertToClusterJobEntity(job)

			if got.Status != s.output {
				t.Errorf("Status = %v, 期待値 %v", got.Status, s.output)
			}
		})
	}
}

// stringPtr はstringへのポインタを返すヘルパー関数
func stringPtr(s string) *string {
	return &s
}
