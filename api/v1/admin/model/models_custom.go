package model

type ImportRecord struct {
	ID               int                    `json:"id"`
	FileName         string                 `json:"fileName"`
	MaterialID       uint                   `json:"materialID"`
	DeviceID         uint                   `json:"deviceID"`
	RowCount         int                    `json:"rowCount"`
	RowFinishedCount int                    `json:"rowFinishedCount"`
	Finished         bool                   `json:"finished"`
	Error            *string                `json:"error"`
	FileSize         int                    `json:"fileSize"`
	UserID           uint                   `json:"userID"`
	ImportType       ImportRecordImportType `json:"importType"`
	DecodeTemplateID uint                   `json:"decodeTemplateID"`
}
