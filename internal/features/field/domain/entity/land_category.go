package entity

// LandCategory は土地種別エンティティ
type LandCategory struct {
	Code        string
	Name        string
	Description *string
}

// NewLandCategory は新しいLandCategoryを作成する
func NewLandCategory(code, name string) *LandCategory {
	return &LandCategory{
		Code: code,
		Name: name,
	}
}
