// Package entity はクラスタリング機能のドメインエンティティを定義する
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Resolution はH3解像度を表す
type Resolution int

const (
	Res3 Resolution = 3 // 約100km - 地方レベル
	Res5 Resolution = 5 // 約10km - 都道府県レベル
	Res7 Resolution = 7 // 約1km - 市区町村レベル
	Res9 Resolution = 9 // 約100m - 詳細レベル
)

// AllResolutions は全てのサポートされている解像度のリスト
var AllResolutions = []Resolution{Res3, Res5, Res7, Res9}

// IsValid は解像度が有効かどうかを判定する
func (r Resolution) IsValid() bool {
	switch r {
	case Res3, Res5, Res7, Res9:
		return true
	default:
		return false
	}
}

// String は解像度の文字列表現を返す
func (r Resolution) String() string {
	switch r {
	case Res3:
		return "res3"
	case Res5:
		return "res5"
	case Res7:
		return "res7"
	case Res9:
		return "res9"
	default:
		return "unknown"
	}
}

// Cluster はH3クラスタリング結果のエンティティ
type Cluster struct {
	ID           uuid.UUID
	Resolution   Resolution
	H3Index      string  // H3インデックス(16進数文字列)
	FieldCount   int32   // クラスターに含まれる圃場数
	CenterLat    float64 // クラスター中心の緯度
	CenterLng    float64 // クラスター中心の経度
	CalculatedAt time.Time
}

// NewCluster は新しいClusterを作成する
func NewCluster(resolution Resolution, h3Index string, fieldCount int32, centerLat, centerLng float64) *Cluster {
	return &Cluster{
		ID:           uuid.New(),
		Resolution:   resolution,
		H3Index:      h3Index,
		FieldCount:   fieldCount,
		CenterLat:    centerLat,
		CenterLng:    centerLng,
		CalculatedAt: time.Now(),
	}
}

// ClusterResult はAPIレスポンス用のクラスター情報
type ClusterResult struct {
	H3Index string
	Lat     float64
	Lng     float64
	Count   int32
}

// ToResult はClusterをClusterResultに変換する
func (c *Cluster) ToResult() *ClusterResult {
	return &ClusterResult{
		H3Index: c.H3Index,
		Lat:     c.CenterLat,
		Lng:     c.CenterLng,
		Count:   c.FieldCount,
	}
}
