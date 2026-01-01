package entity

// IdleLandStatus は遊休農地状況エンティティ
type IdleLandStatus struct {
	Code        string
	Name        string
	Description *string
}

// NewIdleLandStatus は新しいIdleLandStatusを作成する
func NewIdleLandStatus(code, name string) *IdleLandStatus {
	return &IdleLandStatus{
		Code: code,
		Name: name,
	}
}
