package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/gorm"
)

func (r *mutationResolver) Login(ctx context.Context, loginInput model.LoginInput) (*model.User, error) {
	var user orm.User

	if err := orm.DB.Where("username = ? AND password = ?", loginInput.Account, orm.Encrypt(loginInput.Password)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewGQLError("账号或密码不正确", err.Error())
		}

		return nil, NewGQLError("登录失败", err.Error())
	}

	token := logic.GenToken(user.Password)
	if err := orm.DB.Model(&user).Update("access_token", token).Error; err != nil {
		return nil, NewGQLError("登录失败", err.Error())
	}

	gc := logic.GetGinContext(ctx)
	if gc != nil {
		gc.Header("Access-Token", token)
	}

	return &model.User{
		ID:      int(user.ID),
		Account: user.Username,
		Admin:   user.Admin,
	}, nil
}

func (r *mutationResolver) Setting(ctx context.Context, settingInput model.SettingInput) (*model.SystemConfig, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	user := logic.CurrentUser(ctx)
	if user == nil || !user.Admin {
		return nil, NewGQLError("添加系统配置失败，您不是Admin", fmt.Sprintf("%+v", *user))
	}

	conf := orm.GetSystemConfig(settingInput.Key)
	if conf == nil {
		conf = &orm.SystemConfig{
			Key:   settingInput.Key,
			Value: settingInput.Value,
		}
	} else {
		conf.Value = settingInput.Value
	}

	if err := orm.DB.Save(conf).Error; err != nil {
		return nil, NewGQLError("添加系统配置失败", err.Error())
	}

	return &model.SystemConfig{
		ID:        int(conf.ID),
		Key:       conf.Key,
		Value:     conf.Value,
		CreatedAt: conf.CreatedAt,
		UpdatedAt: conf.UpdatedAt,
	}, nil
}

func (r *mutationResolver) AddMaterial(ctx context.Context, materialName string) (*model.AddMaterialResponse, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	if !logic.IsMaterialExist(materialName) {
		return nil, NewGQLError("FTP服务器现在没有该料号的数据。", "IsMaterialExist false")
	}

	material := orm.GetMaterialWithName(materialName)
	if material != nil {
		return nil, NewGQLError("料号已经存在，请确认你的输入。", "find material, can't create another one.")
	}

	m := orm.Material{Name: materialName}
	if err := orm.DB.Create(&m).Error; err != nil {
		return nil, NewGQLError("创建料号失败", err.Error())
	}

	fileIDs, _ := logic.FetchMaterialDatas(m, nil, nil)
	var status = model.FetchStatus{
		Message: "创建料号成功",
		Pending: true,
		FileIDs: fileIDs,
	}
	if len(fileIDs) == 0 {
		status.Pending = false
		status.Message = "已为您创建料号，但FTP服务器暂无该料号最近一个月的数据"
	}

	materialOut := model.Material{
		ID:   m.ID,
		Name: m.Name,
	}

	return &model.AddMaterialResponse{
		Material: &materialOut,
		Status:   &status,
	}, nil
}

func (r *mutationResolver) Active(ctx context.Context, accessToken string) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	user := logic.CurrentUser(ctx)
	if user == nil {
		return nil, NewGQLError("用户未登录", "current user is nil")
	}

	return &model.User{
		ID:      int(user.ID),
		Account: user.Username,
		Admin:   user.Admin,
	}, nil
}

func (r *queryResolver) Products(ctx context.Context, searchInput model.Search, page int, limit int) (*model.ProductWrap, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("差缺数据缺少料号ID", "searchInput.MaterialID is nil")
	}
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "")
	}

	offset := (page - 1) * limit

	var conditions []string
	var vars []interface{}
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("您所查找的料号不存在", fmt.Sprintf("get material with id = %v failed", *searchInput.MaterialID))
	}
	conditions = append(conditions, "material_id = ?")
	vars = append(vars, material.ID)

	if searchInput.DeviceID != nil {
		device := orm.GetDeviceWithID(*searchInput.DeviceID)
		if device != nil {
			conditions = append(conditions, "device_id = ?")
			vars = append(vars, device.ID)
		}
	}

	if searchInput.BeginTime != nil {
		conditions = append(conditions, "created_at > ?")
		vars = append(vars, searchInput.BeginTime)
	}

	if searchInput.EndTime != nil {
		conditions = append(conditions, "created_at < ?")
		vars = append(vars, searchInput.EndTime)
	}

	fmt.Println(conditions)
	cond := strings.Join(conditions, " AND ")
	var products []orm.Product
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("created_at desc").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound { // 数据未找到，需要去FTP拉取
			fileIDs, _ := logic.FetchMaterialDatas(*material, nil, nil)
			status := &model.FetchStatus{FileIDs: fileIDs, Pending: false}
			if len(fileIDs) == 0 {
				status.Message = "抱歉，未能在FTP服务器中查找到此料号在该时间段内的数据"
				return &model.ProductWrap{Status: status}, nil
			}

			status.Message = "需要从FTP服务器获取该时间段内料号数据"
			status.Pending = true
			return &model.ProductWrap{Status: status}, nil
		}

		return nil, NewGQLError("获取数据失败，请重试", err.Error())
	}

	var total int
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Count(&total).Error; err != nil {
		return nil, NewGQLError("统计产品数量失败", err.Error())
	}

	var productUUIDs []string
	for _, p := range products {
		productUUIDs = append(productUUIDs, p.UUID)
	}

	rows, err := orm.DB.Raw(`
	SELECT sv.product_uuid, s.name, sv.value FROM size_values AS sv
	JOIN sizes AS s ON sv.size_id = s.id
	WHERE sv.product_uuid in (?)
	ORDER BY sv.product_uuid, s.index
	`, productUUIDs).Rows()
	if err != nil {
		return nil, NewGQLError("获取产品尺寸数据失败", err.Error())
	}
	defer rows.Close()

	var uuid, name string
	var value float64
	productSizeValueMap := make(map[string]map[string]interface{})
	for rows.Next() {
		rows.Scan(&uuid, &name, &value)
		if p, ok := productSizeValueMap[uuid]; ok {
			p[name] = value
			continue
		}

		productSizeValueMap[uuid] = map[string]interface{}{name: value}
	}

	var outProducts []*model.Product
	for _, p := range products {
		op := &model.Product{
			ID:         p.ID,
			UUID:       p.UUID,
			MaterialID: p.MaterialID,
			DeviceID:   p.DeviceID,
			Qualified:  p.Qualified,
			CreatedAt:  p.CreatedAt,
		}
		if mp, ok := productSizeValueMap[p.UUID]; ok {
			op.SizeValue = mp
		}
		outProducts = append(outProducts, op)
	}

	var sizeNames []string
	orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Order("sizes.index asc").Pluck("name", &sizeNames)

	status := model.FetchStatus{Pending: false}
	return &model.ProductWrap{
		TableHeader: sizeNames,
		Products:    outProducts,
		Status:      &status,
		Total:       total,
	}, nil
}

func (r *queryResolver) AnalyzeSize(ctx context.Context, searchInput model.Search) (*model.SizeResult, error) {
	if searchInput.SizeID == nil {
		return nil, NewGQLError("尺寸ID不能为空", "")
	}
	size := orm.GetSizeWithID(*searchInput.SizeID)
	if size == nil {
		return nil, NewGQLError("没有在数据库找到改尺寸", "")
	}

	var sizeValues []orm.SizeValue
	cond := "size_id = ?"
	vals := []interface{}{*searchInput.SizeID}

	if searchInput.DeviceID != nil {
		cond = cond + "AND device_id = ?"
		vals = append(vals, *searchInput.DeviceID)
	}

	if searchInput.BeginTime != nil {
		cond = cond + "AND created_at > ?"
		vals = append(vals, *searchInput.BeginTime)
	}

	if searchInput.EndTime != nil {
		cond = cond + "AND created_at < ?"
		vals = append(vals, *searchInput.EndTime)
	}

	if err := orm.DB.Model(&orm.SizeValue{}).Where(cond, vals...).Find(&sizeValues).Error; err != nil {
		return nil, NewGQLError("获取数据失败", err.Error())
	}

	total := len(sizeValues)
	if total == 0 {
		return &model.SizeResult{}, nil
	}
	ok := 0
	valueSet := make([]float64, 0)
	for _, v := range sizeValues {
		if v.Qualified {
			valueSet = append(valueSet, v.Value)
			ok++
		}
	}
	s := logic.RMSError(valueSet)
	cp := logic.Cp(size.UpperLimit, size.LowerLimit, s)

	freqs := make([]int, 0)
	values := make([]float64, 0)

	rows, err := orm.DB.Model(&orm.SizeValue{}).Where("size_id = ? and qualified = 1", size.ID).Group("value").Select("COUNT(value) as freq, value").Rows()
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			var freq int
			var value float64
			rows.Scan(&freq, &value)
			freqs = append(freqs, freq)
			values = append(values, value)
		}
	}

	normal := logic.Normal(values, freqs)
	cpk := logic.Cpk(size.UpperLimit, size.LowerLimit, normal, s)

	return &model.SizeResult{
		Total:  total,
		Ok:     ok,
		Ng:     total - ok,
		Cp:     cp,
		Cpk:    cpk,
		Normal: normal,
		Dataset: map[string]interface{}{
			"values": values,
			"freqs":  freqs,
		},
	}, nil
}

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, searchInput model.Search) (*model.MaterialResult, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("料号ID不能为空", "searchInput.ID can't be empty")
	}
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("料号不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.MaterialID))
	}
	beginTime := searchInput.BeginTime
	endTime := searchInput.EndTime
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}
	if beginTime == nil {
		t := endTime.AddDate(0, -1, 0)
		beginTime = &t
	}

	var ok int
	var ng int
	orm.DB.Model(&orm.Product{}).Where(
		"material_id = ? and created_at < ? and created_at > ? and qualified = 1",
		searchInput.MaterialID, endTime, beginTime,
	).Count(&ok)
	orm.DB.Model(&orm.Product{}).Where(
		"material_id = ? and created_at < ? and created_at > ? and qualified = 0",
		searchInput.MaterialID, endTime, beginTime,
	).Count(&ng)
	out := model.Material{
		ID:   material.ID,
		Name: material.Name,
	}

	return &model.MaterialResult{
		Material: &out,
		Ok:       ok,
		Ng:       ng,
	}, nil
}

func (r *queryResolver) AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	if searchInput.DeviceID == nil {
		return nil, NewGQLError("设备ID不能为空", "searchInput.DeviceID can't be empty")
	}
	device := orm.GetDeviceWithID(*searchInput.DeviceID)
	if device == nil {
		return nil, NewGQLError("设备不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.DeviceID))
	}
	beginTime := searchInput.BeginTime
	endTime := searchInput.EndTime
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}
	if beginTime == nil {
		t := endTime.AddDate(0, -1, 0)
		beginTime = &t
	}

	var ok int
	var ng int
	orm.DB.Model(&orm.Product{}).Where(
		"device_id = ? and created_at < ? and created_at > ? and qualified = 1",
		searchInput.DeviceID, endTime, beginTime,
	).Count(&ok)
	orm.DB.Model(&orm.Product{}).Where(
		"device_id = ? and created_at < ? and created_at > ? and qualified = 0",
		searchInput.DeviceID, endTime, beginTime,
	).Count(&ng)
	out := model.Device{
		ID:   device.ID,
		Name: device.Name,
	}

	return &model.DeviceResult{
		Device: &out,
		Ok:     ok,
		Ng:     ng,
	}, nil
}

func (r *queryResolver) Sizes(ctx context.Context, page int, limit int, materialID int) ([]*model.Size, error) {
	var sizes []orm.Size
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Where("material_id = ?", materialID).Order("sizes.index asc").Limit(limit).Offset(offset).Find(&sizes).Error; err != nil {
		return nil, NewGQLError("获取料号信息失败", err.Error())
	}
	var outs []*model.Size
	for _, v := range sizes {
		outs = append(outs, &model.Size{
			ID:         v.ID,
			Name:       v.Name,
			UpperLimit: v.UpperLimit,
			LowerLimit: v.LowerLimit,
		})
	}
	return outs, nil
}

func (r *queryResolver) Materials(ctx context.Context, page int, limit int) ([]*model.Material, error) {
	var materials []orm.Material
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Order("id desc").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		return nil, NewGQLError("获取料号信息失败", err.Error())
	}
	var outs []*model.Material
	for _, v := range materials {
		outs = append(outs, &model.Material{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return outs, nil
}

func (r *queryResolver) Devices(ctx context.Context, page int, limit int, materialID int) ([]*model.Device, error) {
	var devices []orm.Device
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Where("material_id = ?", materialID).Order("id desc").Limit(limit).Offset(offset).Find(&devices).Error; err != nil {
		return nil, NewGQLError("获取设备信息失败", err.Error())
	}
	var outs []*model.Device
	for _, v := range devices {
		outs = append(outs, &model.Device{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return outs, nil
}

func (r *queryResolver) DataFetchFinishPercent(ctx context.Context, fileIDs []*int) (float64, error) {
	total := len(fileIDs)
	if total == 0 {
		return 0, nil
	}
	var finished int
	orm.DB.Model(&orm.FileList{}).Where("id in ? and finished = 1", fileIDs).Count(&finished)

	return float64(finished / total), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
