package orm

import (
	"testing"
)

func TestCURD(t *testing.T) {
	conf := SystemConfig{
		Key:   "hello",
		Value: "world",
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
