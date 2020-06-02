package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
	"path/filepath"
	"regexp"
)

const fileNameDecodePattern = `([\w]*)-([\w]*)-.*-([A|B|w|b]?).xlsx`

func AddMaterial(ctx context.Context, input model.MaterialCreateInput) (*model.Material, error) {
	gc := getGinContext(ctx)
	user := currentUser(gc)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.DB.Begin()
	var material orm.Material
	tx.Model(&material).Where("name = ?", input.Name).First(&material)
	if material.ID != 0 {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeMaterialAlreadyExists, nil)
	}

	material.Name = input.Name
	if input.CustomerCode != nil {
		material.CustomerCode = *input.CustomerCode
	}
	if input.ProjectRemark != nil {
		material.ProjectRemark = *input.CustomerCode
	}
	if err := tx.Create(&material).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCreateFailedError, err, "material")
	}

	pointColumns := make(types.Map)
	for _, pointInput := range input.Points {
		point := orm.Point{
			Name:       pointInput.Name,
			MaterialID: material.ID,
			UpperLimit: pointInput.Usl,
			LowerLimit: pointInput.Lsl,
			Nominal:    pointInput.Nominal,
		}
		if err := tx.Create(&point).Error; err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCreateFailedError, err, "point")
		}
		pointColumns[pointInput.Name] = pointInput.Index
	}

	decodeTemplate := orm.DecodeTemplate{
		Name:         "默认模板",
		MaterialID:   material.ID,
		UserID:       user.ID,
		Description:  "创建料号时自动生成的默认解析模板",
		DataRowIndex: 15,
		PointColumns: pointColumns,
		Default:      true,
	}
	if err := decodeTemplate.GenDefaultProductColumns(); err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCreateFailedError, err, "decode_templates")
	}
	if err := tx.Create(&decodeTemplate).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCreateFailedError, err, "decode_templates")
	}

	tx.Commit()

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "material")
	}

	// 解析FTP服务器指定料号路径下的所有未解析文件
	if err := FetchMaterialData(&material); err != nil {
		return &out, errormap.SendGQLError(gc, errormap.ErrorCodeCreateSuccessButFetchFailed, err)
	}

	return &out, nil
}

type fetchItem struct {
	Device   orm.Device
	FileName string
}

// FetchMaterialData 判断是否需要从FTP拉取数据
// 给定料号，对比数据库中已拉取文件路径，得出是否有需要拉取的文件路径
func FetchMaterialData(material *orm.Material) error {
	var needFetch []fetchItem

	template, err := material.GetDefaultTemplate()
	if err != nil {
		return errormap.NewOrigin("get default decode template for material(id = %v) failed: %v", material.ID, err)
	}

	ftpFileList, err := ftpclient.GetList("./" + material.Name)
	if err != nil {
		return err
	}

	for _, p := range ftpFileList {
		need, deviceRemark := checkFile(material.ID, p)
		if !need {
			continue
		}
		var device orm.Device
		device.CreateIfNotExist(material.ID, deviceRemark)
		needFetch = append(needFetch, fetchItem{device, p})
	}

	if len(needFetch) == 0 {
		return nil
	}

	return fetchMaterialData(*material, needFetch, template)
}

// FetchMaterialData 获取指定文件中的数据
func fetchMaterialData(material orm.Material, files []fetchItem, dt *orm.DecodeTemplate) error {
	for _, f := range files {
		xr := ftpclient.NewXLSXReader(&material, &f.Device, dt)
		path := resolvePath(material.Name, f.FileName)

		importRecord := &orm.ImportRecord{
			FileName:         f.FileName,
			MaterialID:       material.ID,
			DeviceID:         f.Device.ID,
			DecodeTemplateID: dt.ID,
		}
		if err := orm.Create(importRecord).Error; err != nil {
			// TODO: add log
			log.Errorln(err)
			continue
		}

		go func() {
			log.Warn("start read routine with file: %s\n", path)
			err := xr.Read(path)
			if err != nil {
				log.Error("read path(%s) error: %v", path, err)
				return
			}
			importRecord.RowCount = len(xr.DataSet)
			if err := orm.Save(importRecord).Error; err != nil {
				// TODO: add log
				log.Errorln(err)
				return
			}
			xr.Record = importRecord
			ftpclient.PushStore(xr)
		}()
	}

	return nil
}

// checkFile 仅检查文件是否已经被读取到指定料号
func checkFile(materialID uint, fileName string) (bool, string) {
	var importRecord orm.ImportRecord
	// 查找 当前料号的 当前文件名的 已完成的 且 没有处理错误的 文件导入记录，若存在则忽略此文件
	orm.DB.Model(&importRecord).Where(
		"file_name = ? AND material_id = ? AND finished = 1 AND error IS NULL",
		fileName, materialID,
	).First(&importRecord)

	if importRecord.ID != 0 {
		return false, ""
	}

	reg := regexp.MustCompile(fileNameDecodePattern)
	matched := reg.FindStringSubmatch(fileName)
	if len(matched) != 4 {
		return false, ""
	}
	return true, matched[2]
}

func resolvePath(m, path string) string {
	return fmt.Sprintf("./%s/%s", m, filepath.Base(path))
}
