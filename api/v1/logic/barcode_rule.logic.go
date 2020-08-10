package logic

import (
	"context"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"time"
)

type barCodeStatusAnalyzeResult struct {
	Amount int
	Status int
}

func BarCodeStatusAnalyze(ctx context.Context, materialID int, versionID *int, duration []*time.Time) (*model.BarCodeStatusAnalyzeResponse, error) {
	var material orm.Material
	if err := material.Get(uint(materialID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	var version orm.MaterialVersion
	if versionID != nil {
		if err := version.Get(uint(*versionID)); err != nil {
			return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material_version")
		}
	} else {
		currentVersion, err := material.GetCurrentVersion()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeActiveVersionNotFound, err)
		}
		version = *currentVersion
	}

	query := orm.Model(&orm.Product{}).Where("material_id = ? AND material_version_id = ?", material.ID, version.ID)
	query = query.Group("bar_code_status")
	if len(duration) > 0 {
		query = query.Where("created_at > ?", duration[0])
	}
	if len(duration) > 1 {
		query = query.Where("created_at < ?", duration[1])
	}
	query = query.Select("bar_code_status, COUNT(*)")

	var results []barCodeStatusAnalyzeResult
	rows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "product")
	}

	for rows.Next() {
		var result barCodeStatusAnalyzeResult
		if err := rows.Scan(&result.Status, &result.Amount); err != nil {
			log.Errorln(err)
			continue
		}
		results = append(results, result)
	}
	_ = rows.Close()

	var response model.BarCodeStatusAnalyzeResponse
	var total, ok int
	for _, r := range results {
		total = total + r.Amount
		if r.Status == orm.BarCodeStatusSuccess {
			ok = ok + r.Amount
		}
	}
	response.Amount = total
	response.Yield = float64(ok) / float64(total)
	ng := total - ok

	for _, r := range results {
		if r.Status == orm.BarCodeStatusSuccess {
			continue
		}

		response.FailedAmounts = append(response.FailedAmounts, r.Amount)
		if ng != 0 {
			response.FailedYields = append(response.FailedYields, float64(r.Amount)/float64(ng))
		} else {
			response.FailedYields = append(response.FailedYields, 0)
		}
		switch r.Status {
		case orm.BarCodeStatusIllegal:
			response.FailedLabels = append(response.FailedLabels, "Illegal")
		case orm.BarCodeStatusTooShort:
			response.FailedLabels = append(response.FailedLabels, "TooShort")
		case orm.BarCodeStatusReadFail:
			response.FailedLabels = append(response.FailedLabels, "ReadError")
		case orm.BarCodeStatusNoRule:
			response.FailedLabels = append(response.FailedLabels, "NoItems")
		}
	}

	return &response, nil
}
