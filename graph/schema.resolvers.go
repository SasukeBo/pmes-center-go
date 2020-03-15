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

	conf := orm.GetSystemConfigCache(settingInput.Key)
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

	orm.CacheSystemConfig(*conf)

	return &model.SystemConfig{
		ID:        int(conf.ID),
		Key:       conf.Key,
		Value:     conf.Value,
		CreatedAt: conf.CreatedAt,
		UpdatedAt: conf.UpdatedAt,
	}, nil
}

func (r *mutationResolver) AddMaterial(ctx context.Context, materialID string) (bool, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return false, err
	}

	material := orm.GetMaterialWithIDCache(materialID)
	if material != nil {
		return false, NewGQLError("料号已经存在，请确认你的输入。", "find material, can't create another one.")
	}

	m := orm.Material{Name: materialID}
	if err := orm.DB.Create(&m).Error; err != nil {
		return false, NewGQLError("创建料号失败", err.Error())
	}

	if !logic.IsMaterialExist(materialID) {
		return true, NewGQLError("料号创建成功，但是FTP服务器现在没有该料号的数据。", "IsMaterialExist false")
	}

	return true, nil
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
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Cpk(ctx context.Context, cpkInput model.Search) (*model.AnalysisResult, error) {
	panic(fmt.Errorf("not implemented"))
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
