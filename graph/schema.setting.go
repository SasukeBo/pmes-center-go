package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
)

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

	var out model.SystemConfig
	if err := copier.Copy(&out, &conf); err != nil {
		return nil, NewGQLError("数据转换发生错误", err.Error())
	}

	return &out, nil
}
