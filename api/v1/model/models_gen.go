// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type BarCodeStatusAnalyzeResponse struct {
	Yield         float64   `json:"yield"`
	Amount        int       `json:"amount"`
	FailedYields  []float64 `json:"failedYields"`
	FailedAmounts []int     `json:"failedAmounts"`
	FailedLabels  []string  `json:"failedLabels"`
}

type DeviceResult struct {
	Device *Device `json:"device"`
	Ok     int     `json:"ok"`
	Ng     int     `json:"ng"`
}

type EchartsResult struct {
	XAxisData        []string               `json:"xAxisData"`
	SeriesData       map[string]interface{} `json:"seriesData"`
	SeriesAmountData map[string]interface{} `json:"seriesAmountData"`
}

// 获取Echart绘图数据所需的参数
type GraphInput struct {
	TargetID       int                    `json:"targetID"`
	XAxis          Category               `json:"xAxis"`
	YAxis          YAxis                  `json:"yAxis"`
	GroupBy        *Category              `json:"groupBy"`
	Duration       []*time.Time           `json:"duration"`
	Limit          *int                   `json:"limit"`
	Sort           *Sort                  `json:"sort"`
	AttributeXAxis *string                `json:"attributeXAxis"`
	AttributeGroup *string                `json:"attributeGroup"`
	Filters        map[string]interface{} `json:"filters"`
}

type Material struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	YieldScore    float64 `json:"yieldScore"`
	CustomerCode  string  `json:"customerCode"`
	ProjectRemark string  `json:"projectRemark"`
	Ok            int     `json:"ok"`
	Ng            int     `json:"ng"`
}

type MaterialResult struct {
	Material *Material `json:"material"`
	Ok       int       `json:"ok"`
	Ng       int       `json:"ng"`
}

type MaterialVersion struct {
	ID          int     `json:"id"`
	Version     string  `json:"version"`
	Amount      int     `json:"amount"`
	Yield       float64 `json:"yield"`
	Description string  `json:"description"`
}

type MaterialsWrap struct {
	Total     int         `json:"total"`
	Materials []*Material `json:"materials"`
}

type PointListWithYieldResponse struct {
	Total int      `json:"total"`
	List  []*Point `json:"list"`
}

type PointResult struct {
	Total   int                    `json:"total"`
	S       float64                `json:"s"`
	Ok      int                    `json:"ok"`
	Ng      int                    `json:"ng"`
	Cp      float64                `json:"cp"`
	Cpk     float64                `json:"cpk"`
	Avg     float64                `json:"avg"`
	Max     float64                `json:"max"`
	Min     float64                `json:"min"`
	Dataset map[string]interface{} `json:"dataset"`
	Point   *Point                 `json:"point"`
}

type ProductAttribute struct {
	Type   string `json:"type"`
	Prefix string `json:"prefix"`
	Label  string `json:"label"`
	Name   string `json:"name"`
	Token  string `json:"token"`
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

type User struct {
	ID      int    `json:"id"`
	Account string `json:"account"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
	UUID    string `json:"uuid"`
}

type Category string

const (
	CategoryDate      Category = "Date"
	CategoryDevice    Category = "Device"
	CategoryShift     Category = "Shift"
	CategoryAttribute Category = "Attribute"
)

var AllCategory = []Category{
	CategoryDate,
	CategoryDevice,
	CategoryShift,
	CategoryAttribute,
}

func (e Category) IsValid() bool {
	switch e {
	case CategoryDate, CategoryDevice, CategoryShift, CategoryAttribute:
		return true
	}
	return false
}

func (e Category) String() string {
	return string(e)
}

func (e *Category) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Category(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Category", str)
	}
	return nil
}

func (e Category) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Sort string

const (
	SortAsc  Sort = "ASC"
	SortDesc Sort = "DESC"
)

var AllSort = []Sort{
	SortAsc,
	SortDesc,
}

func (e Sort) IsValid() bool {
	switch e {
	case SortAsc, SortDesc:
		return true
	}
	return false
}

func (e Sort) String() string {
	return string(e)
}

func (e *Sort) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Sort(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Sort", str)
	}
	return nil
}

func (e Sort) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type YAxis string

const (
	YAxisYield   YAxis = "Yield"
	YAxisUnYield YAxis = "UnYield"
	YAxisAmount  YAxis = "Amount"
)

var AllYAxis = []YAxis{
	YAxisYield,
	YAxisUnYield,
	YAxisAmount,
}

func (e YAxis) IsValid() bool {
	switch e {
	case YAxisYield, YAxisUnYield, YAxisAmount:
		return true
	}
	return false
}

func (e YAxis) String() string {
	return string(e)
}

func (e *YAxis) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = YAxis(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid YAxis", str)
	}
	return nil
}

func (e YAxis) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
