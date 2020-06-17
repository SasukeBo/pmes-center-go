package model

import "time"

type ImportRecord struct {
	ID                 int                    `json:"id"`
	FileName           string                 `json:"fileName"`
	MaterialID         uint                   `json:"materialID"`
	DeviceID           uint                   `json:"deviceID"`
	RowCount           int                    `json:"rowCount"`
	RowFinishedCount   int                    `json:"rowFinishedCount"`
	Status             ImportStatus           `json:"status"`
	ErrorCode          *string                `json:"errorCode"`
	OriginErrorMessage *string                `json:"originErrorMessage"`
	FileSize           int                    `json:"fileSize"`
	UserID             uint                   `json:"userID"`
	ImportType         ImportRecordImportType `json:"importType"`
	DecodeTemplateID   uint                   `json:"decodeTemplateID"`
	CreatedAt          time.Time              `json:"createdAt"`
}

type DecodeTemplate struct {
	ID                   int                    `json:"id"`
	Name                 string                 `json:"name"`
	MaterialID           uint                   `json:"materialID"`
	UserID               uint                   `json:"userID"`
	Description          string                 `json:"description"`
	DataRowIndex         int                    `json:"dataRowIndex"`
	CreatedAtColumnIndex string                 `json:"createdAtColumnIndex"`
	ProductColumns       []*ProductColumn       `json:"productColumns"`
	PointColumns         map[string]interface{} `json:"pointColumns"`
	Default              bool                   `json:"default"`
	CreatedAt            time.Time              `json:"createdAt"`
	UpdatedAt            time.Time              `json:"updatedAt"`
}

type Device struct {
	ID             int    `json:"id"`
	UUID           string `json:"uuid"`
	Name           string `json:"name"`
	Remark         string `json:"remark"`
	IP             string `json:"ip"`
	MaterialID     uint   `json:"materialID"`
	DeviceSupplier string `json:"deviceSupplier"`
	IsRealtime     bool   `json:"isRealtime"`
	Address        string `json:"address"`
}
