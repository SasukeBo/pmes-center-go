package model

import "time"

type BarCodeRule struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Remark     string         `json:"remark"`
	UserID     uint           `json:"userID"`
	CodeLength int            `json:"codeLength"`
	Items      []*BarCodeItem `json:"items"`
	CreatedAt  time.Time      `json:"createdAt"`
}

type MaterialVersion struct {
	ID          int       `json:"id"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	MaterialID  uint      `json:"material"`
	UserID      uint      `json:"userID"`
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
	RowInvalidCount    int                    `json:"rowInvalidCount"`
	Status             ImportStatus           `json:"status"`
	ErrorCode          *string                `json:"errorCode"`
	OriginErrorMessage *string                `json:"originErrorMessage"`
	FileSize           int                    `json:"fileSize"`
	UserID             uint                   `json:"userID"`
	ImportType         ImportRecordImportType `json:"importType"`
	MaterialVersionID  uint                   `json:"materialVersionID"`
	Blocked            bool                   `json:"blocked"`
	Yield              float64                `json:"yield"`
	CreatedAt          time.Time              `json:"createdAt"`
}

type DecodeTemplate struct {
	ID                   int       `json:"id"`
	MaterialID           uint      `json:"materialID"`
	UserID               uint      `json:"userID"`
	DataRowIndex         int       `json:"dataRowIndex"`
	CreatedAtColumnIndex string    `json:"createdAtColumnIndex"`
	MaterialVersionID    uint      `json:"materialVersionID"`
	BarCodeIndex         string    `json:"barCodeIndex"`
	BarCodeRuleID        uint      `json:"barCodeRuleID"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
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
