package orm

import (
	"fmt"
	"testing"
	"time"
)

func TestDevice_GetWithToken(t *testing.T) {
	material := Material{Name: "test material"}
	Create(&material)
	device := Device{
		Name:       "test device",
		Remark:     "test device",
		MaterialID: material.ID,
	}
	Create(&device)
	var device1, device2, device3 Device
	t1 := time.Now()
	if err := device1.GetWithToken(device.UUID); err != nil {
		t.Fatal(err)
	}
	t2 := time.Now()
	if err := device2.GetWithToken(device.UUID); err != nil {
		t.Fatal(err)
	}
	t3 := time.Now()
	device.Name = "test device updated"
	Save(&device)
	t4 := time.Now()
	if err := device3.GetWithToken(device.UUID); err != nil {
		t.Fatal(err)
	}
	t5 := time.Now()

	fmt.Printf("fetch device without cache: %v\n", t2.Sub(t1))
	fmt.Printf("fetch device with cache: %v\n", t3.Sub(t2))
	fmt.Printf("fetch device after cache flush: %v\n", t5.Sub(t4))
	Exec("delete from materials where 1 = 1")
	Exec("delete from devices where 1 = 1")
}

func TestImportRecord_genKey(t *testing.T) {
	record := ImportRecord{}
	key := record.genKey(1)
	fmt.Println(key)
}
