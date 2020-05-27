// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type AddMaterialResponse struct {
	Material *Material    `json:"material"`
	Status   *FetchStatus `json:"status"`
}

type Device struct {
	ID   *int    `json:"id"`
	Name *string `json:"name"`
}

type DeviceResult struct {
	Device *Device      `json:"device"`
	Ok     *int         `json:"ok"`
	Ng     *int         `json:"ng"`
	Status *FetchStatus `json:"status"`
}

type ExportResponse struct {
	Percent  float64 `json:"percent"`
	Message  string  `json:"message"`
	FileName *string `json:"fileName"`
	Finished bool    `json:"finished"`
}

type LoginInput struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type Material struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	CustomerCode  *string `json:"customerCode"`
	ProjectRemark *string `json:"projectRemark"`
}

type MaterialCreateInput struct {
	Name          string  `json:"name"`
	CustomerCode  *string `json:"customerCode"`
	ProjectRemark *string `json:"projectRemark"`
}

type MaterialResult struct {
	Material *Material    `json:"material"`
	Ok       *int         `json:"ok"`
	Ng       *int         `json:"ng"`
	Status   *FetchStatus `json:"status"`
}

type MaterialUpdateInput struct {
	ID            int     `json:"id"`
	CustomerCode  *string `json:"customerCode"`
	ProjectRemark *string `json:"projectRemark"`
}

type MaterialWrap struct {
	Total     *int        `json:"total"`
	Materials []*Material `json:"materials"`
}

type Point struct {
	ID         *int     `json:"id"`
	Name       *string  `json:"name"`
	UpperLimit *float64 `json:"upperLimit"`
	Norminal   *float64 `json:"norminal"`
	LowerLimit *float64 `json:"lowerLimit"`
}

type PointResult struct {
	Total   *int                   `json:"total"`
	S       *float64               `json:"s"`
	Ok      *int                   `json:"ok"`
	Ng      *int                   `json:"ng"`
	Cp      *float64               `json:"cp"`
	Cpk     *float64               `json:"cpk"`
	Avg     *float64               `json:"avg"`
	Max     *float64               `json:"max"`
	Min     *float64               `json:"min"`
	Dataset map[string]interface{} `json:"dataset"`
	Point   *Point                 `json:"point"`
}

type PointResultsWrap struct {
	PointResults []*PointResult `json:"pointResults"`
	Total        int            `json:"total"`
}

type Product struct {
	ID          *int                   `json:"id"`
	UUID        *string                `json:"uuid"`
	MaterialID  *int                   `json:"materialID"`
	DeviceID    *int                   `json:"deviceID"`
	Qualified   *bool                  `json:"qualified"`
	PointValue  map[string]interface{} `json:"pointValue"`
	CreatedAt   *time.Time             `json:"createdAt"`
	D2code      *string                `json:"d2code"`
	LineID      *string                `json:"lineID"`
	JigID       *string                `json:"jigID"`
	MouldID     *string                `json:"mouldID"`
	ShiftNumber *string                `json:"shiftNumber"`
}

type ProductWrap struct {
	TableHeader []string     `json:"tableHeader"`
	Products    []*Product   `json:"products"`
	Status      *FetchStatus `json:"status"`
	Total       *int         `json:"total"`
}

type Search struct {
	// 料号，指定料号
	MaterialID *int `json:"materialID"`
	// 设备名称，如果不为空则指定该设备生产
	DeviceID *int `json:"deviceID"`
	// 查询时间范围起始时间
	BeginTime *time.Time `json:"beginTime"`
	// 查询时间范围结束时间
	EndTime *time.Time `json:"endTime"`
	// 其他查询条件以map形式传递
	Extra map[string]interface{} `json:"extra"`
}

type SettingInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Size struct {
	ID         *int    `json:"id"`
	Name       *string `json:"name"`
	MaterialID *int    `json:"MaterialID"`
}

type SizeWrap struct {
	Total *int    `json:"total"`
	Sizes []*Size `json:"sizes"`
}

type SystemConfig struct {
	ID        *int       `json:"id"`
	Key       *string    `json:"key"`
	Value     *string    `json:"value"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type User struct {
	ID      *int    `json:"id"`
	Account *string `json:"account"`
	Admin   *bool   `json:"admin"`
}

type YieldWrap struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type FetchStatus struct {
	Message *string `json:"message"`
	Pending *bool   `json:"pending"`
	FileIDs []int   `json:"fileIDs"`
}
