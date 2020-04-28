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

	gc, err := logic.GetGinContext(ctx)
	if err != nil {
		return nil, err
	}

	if gc != nil {
		gc.Header("Access-Token", token)
	}

	userID := int(user.ID)
	return &model.User{
		ID:      &userID,
		Account: &user.Username,
		Admin:   &user.Admin,
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

	confID := int(conf.ID)
	return &model.SystemConfig{
		ID:        &confID,
		Key:       &conf.Key,
		Value:     &conf.Value,
		CreatedAt: &conf.CreatedAt,
		UpdatedAt: &conf.UpdatedAt,
	}, nil
}

func (r *mutationResolver) AddMaterial(ctx context.Context, materialName string) (*model.AddMaterialResponse, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	if materialName == "" {
		return nil, NewGQLError("料号名称不能为空", "material name is empty string")
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

	end := time.Now()
	begin := end.AddDate(0, 0, -30)
	fileIDs, _ := logic.NeedFetch(&m, &begin, &end)
	message := "创建料号成功"
	var status = model.FetchStatus{
		Message: &message,
		Pending: boolP(true),
		FileIDs: fileIDs,
	}
	if len(fileIDs) == 0 {
		status.Pending = boolP(false)
		status.Message = stringP("已为您创建料号，但FTP服务器暂无该料号最近一个月的数据")
	}

	materialOut := model.Material{
		ID:   &m.ID,
		Name: &m.Name,
	}

	return &model.AddMaterialResponse{
		Material: &materialOut,
		Status:   &status,
	}, nil
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
		ID:      intP(int(user.ID)),
		Account: &user.Username,
		Admin:   &user.Admin,
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

	end := searchInput.EndTime
	if end == nil {
		t := time.Now()
		end = &t
	}
	begin := searchInput.BeginTime
	if begin == nil {
		t := end.AddDate(0, -1, 0)
		begin = &t
	}

	fileIDs, err := logic.NeedFetch(material, begin, end)
	if err != nil {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(false), Message: stringP(err.Error())}
		return &model.ProductWrap{Status: status}, nil
	}

	if len(fileIDs) > 0 {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
		return &model.ProductWrap{Status: status}, nil
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
	conditions = append(conditions, "created_at < ?")
	vars = append(vars, end)
	conditions = append(conditions, "created_at > ?")
	vars = append(vars, begin)

	fmt.Println(conditions)
	cond := strings.Join(conditions, " AND ")
	var products []orm.Product
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("created_at desc").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound { // 无数据
			return &model.ProductWrap{
				TableHeader: nil,
				Products:    nil,
				Status:      nil,
				Total:       intP(0),
			}, nil
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
	WHERE sv.product_uuid IN (?)
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
	for _, i := range products {
		p := i
		op := &model.Product{
			ID:         &p.ID,
			UUID:       &p.UUID,
			MaterialID: &p.MaterialID,
			DeviceID:   &p.DeviceID,
			Qualified:  &p.Qualified,
			CreatedAt:  &p.CreatedAt,
		}
		if mp, ok := productSizeValueMap[p.UUID]; ok {
			op.SizeValue = mp
		}
		outProducts = append(outProducts, op)
	}

	var sizeNames []string
	orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Order("sizes.index asc").Pluck("name", &sizeNames)

	status := model.FetchStatus{Pending: boolP(false)}
	return &model.ProductWrap{
		TableHeader: sizeNames,
		Products:    outProducts,
		Status:      &status,
		Total:       &total,
	}, nil
}

func (r *queryResolver) AnalyzeSize(ctx context.Context, searchInput model.Search) (*model.SizeResult, error) {
	if searchInput.SizeID == nil {
		return nil, NewGQLError("尺寸ID不能为空", "")
	}
	size := orm.GetSizeWithID(*searchInput.SizeID)
	if size == nil {
		return nil, NewGQLError("没有在数据库找到该尺寸", "")
	}

	material := orm.GetMaterialWithID(size.MaterialID)
	if material == nil {
		return nil, NewGQLError("没有找到该尺寸所属的料号", "")
	}

	end := searchInput.EndTime
	if end == nil {
		t := time.Now()
		end = &t
	}
	begin := searchInput.BeginTime
	if begin == nil {
		t := end.AddDate(0, -1, 0)
		begin = &t
	}

	fileIDs, err := logic.NeedFetch(material, begin, end)
	if err != nil {
		return nil, err
	}
	if len(fileIDs) > 0 {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内尺寸数据")}
		return &model.SizeResult{Status: status}, nil
	}

	var conds []string
	var vars []interface{}

	conds = append(conds, "size_id = ?")
	vars = append(vars, *searchInput.SizeID)

	if searchInput.DeviceID != nil {
		conds = append(conds, "device_id = ?")
		vars = append(vars, *searchInput.DeviceID)
	}

	conds = append(conds, "created_at > ?")
	vars = append(vars, *begin)
	conds = append(conds, "created_at < ?")
	vars = append(vars, *end)

	var sizeValues []orm.SizeValue
	cond := strings.Join(conds, " AND ")
	if err := orm.DB.Model(&orm.SizeValue{}).Where(cond, vars...).Find(&sizeValues).Error; err != nil {
		return nil, NewGQLError("获取数据失败", err.Error())
	}

	total := len(sizeValues)
	ok := 0
	valueSet := make([]float64, 0)
	for _, v := range sizeValues {
		if size.Norminal > 0 && v.Value > size.Norminal*100 {
			continue
		}
		valueSet = append(valueSet, v.Value)
		if v.Qualified {
			ok++
		}
	}
	if size.ID == 2861 {
		fmt.Println(valueSet)
	}
	s := logic.RMSError(valueSet)
	cp := logic.Cp(size.UpperLimit, size.LowerLimit, s)

	freqs := make([]int, 0)
	values := make([]float64, 0)
	rows, err := orm.DB.Model(&orm.SizeValue{}).Where(strings.Join(conds, " AND "), vars...).Group("value").Select("COUNT(value) as freq, value").Rows()
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			var freq int
			var value float64
			rows.Scan(&freq, &value)
			if size.Norminal > 0 && value > size.Norminal*100 {
				continue
			}
			values = append(values, value)
			freqs = append(freqs, freq)
		}
	}

	avg := logic.Average(valueSet)
	cpk := logic.Cpk(size.UpperLimit, size.LowerLimit, avg, s)

	return &model.SizeResult{
		Total: &total,
		Ok:    &ok,
		Ng:    intP(total - ok),
		Cp:    &cp,
		Cpk:   &cpk,
		Avg:   &avg,
		Dataset: map[string]interface{}{
			"values": values,
			"freqs":  freqs,
		},
		Status: &model.FetchStatus{Pending: boolP(false)},
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

	fileIDs, err := logic.NeedFetch(material, beginTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(fileIDs) > 0 {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
		return &model.MaterialResult{Status: status}, nil
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
		ID:   &material.ID,
		Name: &material.Name,
	}

	return &model.MaterialResult{
		Material: &out,
		Ok:       &ok,
		Ng:       &ng,
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
	material := orm.GetMaterialWithID(device.MaterialID)
	if material == nil {
		return nil, NewGQLError("设备生产的料号不存在", fmt.Sprintf("get material with id = %v failed", device.MaterialID))
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

	out := model.Device{
		ID:   &device.ID,
		Name: &device.Name,
	}
	fileIDs, err := logic.NeedFetch(material, beginTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(fileIDs) > 0 {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内设备数据")}
		return &model.DeviceResult{Status: status, Device: &out}, nil
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

	return &model.DeviceResult{
		Device: &out,
		Ok:     &ok,
		Ng:     &ng,
	}, nil
}

func (r *queryResolver) Sizes(ctx context.Context, page int, limit int, materialID int) (*model.SizeWrap, error) {
	var sizes []orm.Size
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Where("material_id = ?", materialID).Order("sizes.index asc").Limit(limit).Offset(offset).Find(&sizes).Error; err != nil {
		return nil, NewGQLError("获取尺寸信息失败", err.Error())
	}
	var outs []*model.Size
	for _, v := range sizes {
		s := v
		outs = append(outs, &model.Size{
			ID:         &s.ID,
			Name:       &s.Name,
			UpperLimit: &s.UpperLimit,
			Norminal:   &s.Norminal,
			LowerLimit: &s.LowerLimit,
		})
	}
	var count int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", materialID).Count(&count).Error; err != nil {
		return nil, NewGQLError("统计尺寸数量失败", err.Error())
	}

	return &model.SizeWrap{
		Total: &count,
		Sizes: outs,
	}, nil
}

func (r *queryResolver) Materials(ctx context.Context, page int, limit int) (*model.MaterialWrap, error) {
	var materials []orm.Material
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Order("id desc").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		return nil, NewGQLError("获取料号信息失败", err.Error())
	}
	var outs []*model.Material
	for _, i := range materials {
		v := i
		outs = append(outs, &model.Material{
			ID:   &v.ID,
			Name: &v.Name,
		})
	}
	var count int
	if err := orm.DB.Model(&orm.Material{}).Count(&count).Error; err != nil {
		return nil, NewGQLError("统计料号数量失败", err.Error())
	}
	return &model.MaterialWrap{
		Total:     &count,
		Materials: outs,
	}, nil
}

func (r *queryResolver) Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	var devices []orm.Device
	if err := orm.DB.Where("material_id = ?", materialID).Find(&devices).Error; err != nil {
		return nil, NewGQLError("获取设备信息失败", err.Error())
	}
	var outs []*model.Device
	for _, i := range devices {
		v := i
		outs = append(outs, &model.Device{
			ID:   &v.ID,
			Name: &v.Name,
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
	orm.DB.Model(&orm.FileList{}).Where("id in (?) and finished = 1", fileIDs).Count(&finished)

	return float64(finished / total), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func stringP(s string) *string {
	return &s
}
func boolP(b bool) *bool {
	return &b
}
func intP(i int) *int {
	return &i
}
