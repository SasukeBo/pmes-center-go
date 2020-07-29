package orm

const (
	BarCodeDecodeFunc1 = "D5X_PRL_2D_BARCODE"
	BarCodeDecodeFunc2 = "CNC_LINE_TRACEABILITY_CODE_FOR_I89"
	BarCodeDecodeFunc3 = "CNC_LINE_TRACEABILITY_CODE_FOR_I94"
)

type BarCodeDecodeFunc struct {
	ID     uint   `gorm:"primary_key;column:id"`
	Name   string `gorm:"not null;unique_index"` // 函数名
	Remark string `gorm:"not null"`              // 函数描述
}

func setupBarCodeDecodeFuncs() {
	Create(&BarCodeDecodeFunc{
		Name:   BarCodeDecodeFunc1,
		Remark: "D5X PRL 2D Barcode",
	})
	Create(&BarCodeDecodeFunc{
		Name:   BarCodeDecodeFunc2,
		Remark: "CNC line traceability code for I89",
	})
	Create(&BarCodeDecodeFunc{
		Name:   BarCodeDecodeFunc3,
		Remark: "CNC line traceability code for I94",
	})
}
