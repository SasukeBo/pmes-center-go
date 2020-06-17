package admin

import (
	"github.com/SasukeBo/ftpviewer/errormap"
	test "github.com/SasukeBo/ftpviewer/intergration_test"
	"github.com/SasukeBo/ftpviewer/orm"
	"testing"
)

func TestImportRecord(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)

	t.Run("Test get list of import_records", func(t *testing.T) {
		material := orm.Material{
			Name:          "test_material",
			CustomerCode:  "test_material_customer_code",
			ProjectRemark: "test_material_project_remark",
		}
		orm.Create(&material)
		template := orm.DecodeTemplate{
			Name:       "test_template",
			MaterialID: material.ID,
			UserID:     test.Data.Admin.ID,
		}
		orm.Create(&template)
		device := orm.Device{
			Name:       "test_device",
			Remark:     "test_device_remark",
			MaterialID: material.ID,
		}
		orm.Create(&device)
		record1 := orm.ImportRecord{
			FileName:           "record1_file_name",
			Path:               "record1_file_path",
			MaterialID:         material.ID,
			DeviceID:           device.ID,
			RowCount:           100,
			RowFinishedCount:   100,
			Status:             orm.ImportStatusFailed,
			ErrorCode:          errormap.ErrorCodeImportFailedWithPanic,
			OriginErrorMessage: "unknown error",
			UserID:             test.Data.User.ID,
			DecodeTemplateID:   template.ID,
		}
		orm.Create(&record1)
		record2 := orm.ImportRecord{
			FileName:         "record2_file_name",
			Path:             "record2_file_path",
			MaterialID:       material.ID,
			DeviceID:         device.ID,
			RowCount:         100,
			RowFinishedCount: 100,
			Status:           orm.ImportStatusFinished,
			UserID:           test.Data.Admin.ID,
			DecodeTemplateID: template.ID,
		}
		orm.Create(&record2)

		tester.SetHeader("Lang", errormap.ZH_CN)
		ret := tester.API1Admin(listImportRecordsGQL, test.Object{
			"materialID": material.ID,
			"deviceID":   device.ID,
			"page":       1,
			"limit":      1,
		}).GQLObject().Path("$.data.response")
		ret.Object().Value("total").Equal(2)
		arr := ret.Object().Value("importRecords").Array()
		arr.Length().Equal(1)
		record := arr.First().Object()
		record.Path("$.material.id").Equal(material.ID)
		record.Path("$.device.id").Equal(device.ID)
		record.Path("$.user.id").Equal(test.Data.User.ID)
		record.Path("$.decodeTemplate.id").Equal(template.ID)
	})
}
