package test

import (
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"testing"
	"time"
)

// NOTE: this test need your ftp service working
// - docker-compose up ftp
// - then create a directory named 1828
// - put a data file into this directory
func TestMaterial(t *testing.T) {
	tester := newTester(t)
	login(testAdminAccount, testAdminPasswd, true)
	go ftpclient.FTPWorker()
	tester.API1(createMaterialGQL, object{
		"input": object{
			"name":          "1828",
			"customerCode":  "613-12760",
			"projectRemark": "D53 PRL TOP",
			"points": []object{
				{
					"name":    "FAI3_G7",
					"usl":     5.36,
					"nominal": 5.31,
					"lsl":     5.26,
					"index":   7,
				},
				{
					"name":    "FAI3_G8",
					"usl":     5.36,
					"nominal": 5.31,
					"lsl":     5.26,
					"index":   8,
				},
				{
					"name":    "FAI4_G1",
					"usl":     4.28,
					"nominal": 4.23,
					"lsl":     4.18,
					"index":   9,
				},
				{
					"name":    "FAI4_G2",
					"usl":     4.28,
					"nominal": 4.23,
					"lsl":     4.18,
					"index":   10,
				},
				{
					"name":    "FAI4_G3",
					"usl":     4.28,
					"nominal": 4.23,
					"lsl":     4.18,
					"index":   11,
				},
			},
		},
	}).GQLObject().Path("$.data.response")
	// wait for worker
	<-time.After(4 * time.Second)
}
