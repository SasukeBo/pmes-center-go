package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

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

func (r *mutationResolver) AddMaterial(ctx context.Context, materialName string) (string, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return "error", err
	}

	if !logic.IsMaterialExist(materialName) {
		return "error", NewGQLError("FTP服务器现在没有该料号的数据。", "IsMaterialExist false")
	}

	material := orm.GetMaterialWithName(materialName)
	if material != nil {
		return "error", NewGQLError("料号已经存在，请确认你的输入。", "find material, can't create another one.")
	}

	m := orm.Material{Name: materialName}
	if err := orm.DB.Create(&m).Error; err != nil {
		return "error", NewGQLError("创建料号失败", err.Error())
	}

	if err := logic.FetchMaterialDatas(m, nil, nil); err != nil {
		return "error", NewGQLError(err.Error(), fmt.Sprintf("logic.FetchMaterialDatas(%s, nil, nil)", materialName))
	}

	return "success", nil
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

func (r *queryResolver) Products(ctx context.Context, searchInput model.Search) (*model.ProductWrap, error) {
	cond := "WHERE (1=1)"
	vals := make([]interface{}, 0)
	var products []orm.Product
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material != nil {
		cond = cond + "AND material_id = ?"
		vals = append(vals, material.ID)
	}

	device := orm.GetDeviceWithID(*searchInput.DeviceID)
	if device != nil {
		cond = cond + "AND device_id = ?"
		vals = append(vals, device.ID)
	}

	if searchInput.BeginTime != nil {
		cond = cond + "AND producted_at > ?"
		vals = append(vals, searchInput.BeginTime)
	}

	if searchInput.EndTime != nil {
		cond = cond + "AND producted_at < ?"
		vals = append(vals, searchInput.EndTime)
	}

	fmt.Println(cond)
	if err := orm.DB.Where(cond, vals...).Find(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if material == nil {
				return nil, NewGQLError("没有找到产品数据，请确认FTP服务器是否有数据文件", err.Error())
			}

			err := logic.FetchMaterialDatas(*material, searchInput.BeginTime, searchInput.EndTime)
			if err != nil {
				return nil, NewGQLError(err.Error(), fmt.Sprintf("logic.FetchMaterialDatas(%s, nil, nil)", material.Name))
			}
		}

		return nil, NewGQLError("获取数据失败，请重试", err.Error())
	}

	return nil, nil
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
		Total:   total,
		Ok:      ok,
		Ng:      total - ok,
		Cp:      cp,
		Cpk:     cpk,
		Normal:  normal,
		Dataset: map[string]interface{}{
			"values": values,
			"freqs": freqs,
		},
	}, nil
}

func (r *queryResolver) Materials(ctx context.Context, searchInput model.Search) ([]*model.AnalysisResult, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Devices(ctx context.Context, searchInput model.Search) ([]*model.AnalysisResult, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
