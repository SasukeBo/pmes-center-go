package orm

import (
	"testing"

	"github.com/SasukeBo/ftpviewer/test"
)

func TestCURD(t *testing.T) {
	defer test.ClearDB()
	conf := SystemConfig{
		Key:   "名字",
		Value: "汪波",
	}

	if err := DB.Create(&conf).Error; err != nil {
		t.Fatal(err)
	}

	var confr SystemConfig
	if err := DB.Where("id = ?", conf.ID).Find(&confr).Error; err != nil {
		t.Fatal(err)
	}
	if confr.Key != conf.Key {
		t.Fatal("result not match")
	}

	confr.Value = "sasuke"
	if err := DB.Model(&confr).Update(confr).Error; err != nil {
		t.Fatal(err)
	}
	if confr.Value != "sasuke" {
		t.Fatal("result not match")
	}

	if err := DB.Delete(&confr).Error; err != nil {
		t.Fatal(err)
	}
}

func TestExec(t *testing.T) {
	err := DB.Exec("INSERT INTO devices (name, material_id) VALUES (?, ?)", []byte("设备1"), "1765").Error
	if err != nil {
		t.Fatal(err)
	}
}
