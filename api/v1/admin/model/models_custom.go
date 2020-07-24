package model

import "time"

type MaterialVersion struct {
	ID          int       `json:"id"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	MaterialID  uint      `json:"material"`
	UserID      uint      `json:"user"`
	Active      bool      `json:"active"`
	Amount      int       `json:"amount"`
	Yield       float64   `json:"yield"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ImportRecord struct {
	ID                 int                    `json:"id"`
	FileID             *uint                  `json:"fileID"`
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
	Blocked            bool                   `json:"blocked"`
	Yield              float64                `json:"yield"`
	CreatedAt          time.Time              `json:"createdAt"`
}

type DecodeTemplate struct {
	ID                   int              `json:"id"`
	MaterialID           uint             `json:"materialID"`
	UserID               uint             `json:"userID"`
	DataRowIndex         int              `json:"dataRowIndex"`
	CreatedAtColumnIndex string           `json:"createdAtColumnIndex"`
	ProductColumns       []*ProductColumn `json:"productColumns"`
	MaterialVersionID    uint             `json:"materialVersionID"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
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

type File struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Token       string `json:"token"`
	UserID      uint   `json:"userID"`
	Size        int    `json:"size"`
	ContentType string `json:"contentType"`
}
