package logic

import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestHandleExport(t *testing.T) {
	opID := uuid.New().String()
	now := time.Now()
	begin := now.AddDate(-1, 0, 0)
	go HandleExport(opID, &orm.Material{ID: 12, Name: "1828"}, model.Search{BeginTime: &begin, EndTime: &now}, "material_id = ?", 12)
	<-time.After(time.Second)
	rsp := handlerCache[opID]
	if rsp == nil {
		t.Fatal("response is nil")
		return
	}
	for {
		<-time.After(time.Second)
		fmt.Printf("message: %s, err: %v, percent: %v, path: %v\n", rsp.message, rsp.err, rsp.percent, rsp.fileName)
		if rsp.err != nil || rsp.finished {
			break
		}
	}
}
