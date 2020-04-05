// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type AnalysisResult struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Cp      float64                `json:"cp"`
	Cpk     float64                `json:"cpk"`
	Ok      int                    `json:"ok"`
	Ng      int                    `json:"ng"`
	Normal  float64                `json:"normal"`
	Dataset map[string]interface{} `json:"dataset"`
}

type LoginInput struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type Material struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MaterialResult struct {
	Material *Material              `json:"material"`
	Ok       int                    `json:"ok"`
	Ng       int                    `json:"ng"`
	Dataset  map[string]interface{} `json:"dataset"`
}

type Product struct {
	ID         int                    `json:"id"`
	UUID       string                 `json:"uuid"`
	MaterialID string                 `json:"materialID"`
	DeviceID   int                    `json:"deviceID"`
	Qualified  bool                   `json:"qualified"`
	SizeValue  map[string]interface{} `json:"sizeValue"`
	CreatedAt  time.Time              `json:"createdAt"`
}

type ProductWrap struct {
	TableHeader   []*string  `json:"tableHeader"`
	Products      []*Product `json:"products"`
	Count         int        `json:"count"`
	QualifiedRate float64    `json:"qualifiedRate"`
}

type Search struct {
	// 料号，指定料号
	MaterialID *int `json:"materialID"`
	// 设备名称，如果不为空则指定该设备生产
	DeviceID *int `json:"deviceID"`
	// 尺寸，如果不为空则指定改尺寸数据
	SizeID *int `json:"sizeID"`
	// 查询时间范围起始时间
	BeginTime *time.Time `json:"beginTime"`
	// 查询时间范围结束时间
	EndTime *time.Time `json:"endTime"`
}

type SettingInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SizeResult struct {
	Total   int                    `json:"total"`
	Ok      int                    `json:"ok"`
	Ng      int                    `json:"ng"`
	Cp      float64                `json:"cp"`
	Cpk     float64                `json:"cpk"`
	Normal  float64                `json:"normal"`
	Dataset map[string]interface{} `json:"dataset"`
}

type SystemConfig struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	ID      int    `json:"id"`
	Account string `json:"account"`
	Admin   bool   `json:"admin"`
}
