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

type DecodeTemplateInput struct {
	ID                   *int                   `json:"id"`
	Name                 string                 `json:"name"`
	MaterialID           int                    `json:"materialID"`
	Description          *string                `json:"description"`
	DataRowIndex         int                    `json:"dataRowIndex"`
	CreatedAtColumnIndex string                 `json:"createdAtColumnIndex"`
	ProductColumns       []*ProductColumnInput  `json:"productColumns"`
	PointColumns         map[string]interface{} `json:"pointColumns"`
	Default              bool                   `json:"default"`
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
}

type PointCreateInput struct {
	ID         *int    `json:"id"`
	Name       string  `json:"name"`
	UpperLimit float64 `json:"upperLimit"`
	Nominal    float64 `json:"nominal"`
	LowerLimit float64 `json:"lowerLimit"`
}

type ProductColumn struct {
	Prefix string            `json:"prefix"`
	Label  string            `json:"label"`
	Token  string            `json:"token"`
	Index  string            `json:"index"`
	Type   ProductColumnType `json:"type"`
}

type ProductColumnInput struct {
	Prefix string            `json:"prefix"`
	Token  string            `json:"token"`
	Label  string            `json:"label"`
	Index  string            `json:"index"`
	Type   ProductColumnType `json:"type"`
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

type ProductColumnType string

const (
	ProductColumnTypeString   ProductColumnType = "String"
	ProductColumnTypeInteger  ProductColumnType = "Integer"
	ProductColumnTypeFloat    ProductColumnType = "Float"
	ProductColumnTypeDatetime ProductColumnType = "Datetime"
)

var AllProductColumnType = []ProductColumnType{
	ProductColumnTypeString,
	ProductColumnTypeInteger,
	ProductColumnTypeFloat,
	ProductColumnTypeDatetime,
}

func (e ProductColumnType) IsValid() bool {
	switch e {
	case ProductColumnTypeString, ProductColumnTypeInteger, ProductColumnTypeFloat, ProductColumnTypeDatetime:
		return true
	}
	return false
}

func (e ProductColumnType) String() string {
	return string(e)
}

func (e *ProductColumnType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ProductColumnType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ProductColumnType", str)
	}
	return nil
}

func (e ProductColumnType) MarshalGQL(w io.Writer) {
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
