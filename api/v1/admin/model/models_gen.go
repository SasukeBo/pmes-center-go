// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type AddUserInput struct {
	Name     string `json:"name"`
	Account  string `json:"account"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"isAdmin"`
}

type BarCodeItem struct {
	Label           string          `json:"label"`
	Key             string          `json:"key"`
	IndexRange      []int           `json:"indexRange"`
	Type            BarCodeItemType `json:"type"`
	DayCode         []string        `json:"dayCode"`
	DayCodeReject   []string        `json:"dayCodeReject"`
	MonthCode       []string        `json:"monthCode"`
	MonthCodeReject []string        `json:"monthCodeReject"`
	CategorySet     []string        `json:"categorySet"`
}

type BarCodeItemInput struct {
	Label           string          `json:"label"`
	Key             string          `json:"key"`
	IndexRange      []int           `json:"indexRange"`
	Type            BarCodeItemType `json:"type"`
	DayCode         []string        `json:"dayCode"`
	MonthCode       []string        `json:"monthCode"`
	DayCodeReject   []string        `json:"dayCodeReject"`
	MonthCodeReject []string        `json:"monthCodeReject"`
	CategorySet     []string        `json:"categorySet"`
}

type BarCodeRuleInput struct {
	ID         *int                `json:"id"`
	Name       string              `json:"name"`
	Remark     string              `json:"remark"`
	CodeLength int                 `json:"codeLength"`
	Items      []*BarCodeItemInput `json:"items"`
}

type BarCodeRuleWrap struct {
	Total int            `json:"total"`
	Rules []*BarCodeRule `json:"rules"`
}

type DecodeTemplateInput struct {
	ID                   int                 `json:"id"`
	DataRowIndex         int                 `json:"dataRowIndex"`
	BarCodeIndex         *string             `json:"barCodeIndex"`
	BarCodeRuleID        *int                `json:"barCodeRuleID"`
	CreatedAtColumnIndex string              `json:"createdAtColumnIndex"`
	PointColumns         []*PointColumnInput `json:"pointColumns"`
}

type DeviceInput struct {
	ID             *int    `json:"id"`
	Name           string  `json:"name"`
	Remark         string  `json:"remark"`
	IP             *string `json:"ip"`
	MaterialID     int     `json:"materialID"`
	DeviceSupplier *string `json:"deviceSupplier"`
	IsRealtime     bool    `json:"isRealtime"`
	Address        *string `json:"address"`
}

type DeviceWrap struct {
	Total   int       `json:"total"`
	Devices []*Device `json:"devices"`
}

type ImportRecordSearch struct {
	Date     *time.Time      `json:"date"`
	Duration []*time.Time    `json:"duration"`
	Status   []*ImportStatus `json:"status"`
	FileName *string         `json:"fileName"`
	UserID   *int            `json:"userID"`
}

type ImportRecordsWrap struct {
	Total         int             `json:"total"`
	ImportRecords []*ImportRecord `json:"importRecords"`
}

type ImportStatusResponse struct {
	Yield            float64      `json:"yield"`
	Status           ImportStatus `json:"status"`
	FileSize         int          `json:"fileSize"`
	RowCount         int          `json:"rowCount"`
	FinishedRowCount int          `json:"finishedRowCount"`
}

type Material struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	YieldScore    float64   `json:"yieldScore"`
	CustomerCode  string    `json:"customerCode"`
	ProjectRemark string    `json:"projectRemark"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type MaterialCreateInput struct {
	Name          string              `json:"name"`
	YieldScore    *float64            `json:"yieldScore"`
	CustomerCode  *string             `json:"customerCode"`
	ProjectRemark *string             `json:"projectRemark"`
	Points        []*PointCreateInput `json:"points"`
}

type MaterialUpdateInput struct {
	ID            int      `json:"id"`
	YieldScore    *float64 `json:"yieldScore"`
	CustomerCode  *string  `json:"customerCode"`
	ProjectRemark *string  `json:"projectRemark"`
}

type MaterialVersionInput struct {
	MaterialID  int                 `json:"materialID"`
	Version     string              `json:"version"`
	Description *string             `json:"description"`
	Active      *bool               `json:"active"`
	Points      []*PointCreateInput `json:"points"`
}

type MaterialVersionUpdateInput struct {
	Version     *string `json:"version"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
}

type MaterialWrap struct {
	Total     int         `json:"total"`
	Materials []*Material `json:"materials"`
}

type Point struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	UpperLimit float64 `json:"upperLimit"`
	Nominal    float64 `json:"nominal"`
	LowerLimit float64 `json:"lowerLimit"`
	Index      string  `json:"index"`
}

type PointColumnInput struct {
	ID    int    `json:"id"`
	Index string `json:"index"`
}

type PointCreateInput struct {
	ID         *int    `json:"id"`
	Name       string  `json:"name"`
	UpperLimit float64 `json:"upperLimit"`
	Nominal    float64 `json:"nominal"`
	LowerLimit float64 `json:"lowerLimit"`
	Index      string  `json:"index"`
}

type SettingInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SystemConfig struct {
	ID        *int       `json:"id"`
	Key       *string    `json:"key"`
	Value     *string    `json:"value"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type User struct {
	ID      int    `json:"id"`
	Account string `json:"account"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
	UUID    string `json:"uuid"`
}

type BarCodeItemType string

const (
	BarCodeItemTypeCategory BarCodeItemType = "Category"
	BarCodeItemTypeDatetime BarCodeItemType = "Datetime"
	BarCodeItemTypeWeekday  BarCodeItemType = "Weekday"
)

var AllBarCodeItemType = []BarCodeItemType{
	BarCodeItemTypeCategory,
	BarCodeItemTypeDatetime,
	BarCodeItemTypeWeekday,
}

func (e BarCodeItemType) IsValid() bool {
	switch e {
	case BarCodeItemTypeCategory, BarCodeItemTypeDatetime, BarCodeItemTypeWeekday:
		return true
	}
	return false
}

func (e BarCodeItemType) String() string {
	return string(e)
}

func (e *BarCodeItemType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = BarCodeItemType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid BarCodeItemType", str)
	}
	return nil
}

func (e BarCodeItemType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ImportRecordImportType string

const (
	ImportRecordImportTypeSystem ImportRecordImportType = "SYSTEM"
	ImportRecordImportTypeUser   ImportRecordImportType = "USER"
)

var AllImportRecordImportType = []ImportRecordImportType{
	ImportRecordImportTypeSystem,
	ImportRecordImportTypeUser,
}

func (e ImportRecordImportType) IsValid() bool {
	switch e {
	case ImportRecordImportTypeSystem, ImportRecordImportTypeUser:
		return true
	}
	return false
}

func (e ImportRecordImportType) String() string {
	return string(e)
}

func (e *ImportRecordImportType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ImportRecordImportType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ImportRecordImportType", str)
	}
	return nil
}

func (e ImportRecordImportType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ImportStatus string

const (
	ImportStatusImporting ImportStatus = "Importing"
	ImportStatusLoading   ImportStatus = "Loading"
	ImportStatusFinished  ImportStatus = "Finished"
	ImportStatusFailed    ImportStatus = "Failed"
	ImportStatusReverted  ImportStatus = "Reverted"
)

var AllImportStatus = []ImportStatus{
	ImportStatusImporting,
	ImportStatusLoading,
	ImportStatusFinished,
	ImportStatusFailed,
	ImportStatusReverted,
}

func (e ImportStatus) IsValid() bool {
	switch e {
	case ImportStatusImporting, ImportStatusLoading, ImportStatusFinished, ImportStatusFailed, ImportStatusReverted:
		return true
	}
	return false
}

func (e ImportStatus) String() string {
	return string(e)
}

func (e *ImportStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ImportStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ImportStatus", str)
	}
	return nil
}

func (e ImportStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ResponseStatus string

const (
	ResponseStatusOk    ResponseStatus = "OK"
	ResponseStatusError ResponseStatus = "ERROR"
)

var AllResponseStatus = []ResponseStatus{
	ResponseStatusOk,
	ResponseStatusError,
}

func (e ResponseStatus) IsValid() bool {
	switch e {
	case ResponseStatusOk, ResponseStatusError:
		return true
	}
	return false
}

func (e ResponseStatus) String() string {
	return string(e)
}

func (e *ResponseStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ResponseStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ResponseStatus", str)
	}
	return nil
}

func (e ResponseStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
