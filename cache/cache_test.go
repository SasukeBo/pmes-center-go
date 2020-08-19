package cache

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/util"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Run("Test Get Set", func(t *testing.T) {
		var key = "my_name_is"
		var value = "sasuke"

		err := Set(key, value)
		if err != nil {
			t.Fatal(err)
		}

		v, err := Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if v != value {
			t.Fatal("value not equal")
		}
	})

	t.Run("Test get bool", func(t *testing.T) {
		var key = "boolkey"
		var value = false
		err := Set(key, value)
		if err != nil {
			t.Fatal(err)
		}
		v, err := GetBool(key)
		if err != nil {
			t.Fatal(err)
		}
		if v != value {
			t.Fatal("value not equal")
		}
	})

	t.Run("Test get float64", func(t *testing.T) {
		var key = "float64key"
		var value = 3.1415926
		err := Set(key, value)
		if err != nil {
			t.Fatal(err)
		}
		v, err := GetFloat(key)
		if err != nil {
			t.Fatal(err)
		}
		if v != value {
			t.Fatal("value not equal")
		}
	})

	t.Run("Test get int", func(t *testing.T) {
		var key = "intkey"
		var value = 1024
		err := Set(key, value)
		if err != nil {
			t.Fatal(err)
		}
		v, err := GetInt(key)
		if err != nil {
			t.Fatal(err)
		}
		if v != value {
			t.Fatal("value not equal")
		}
	})

	t.Run("test set get struct", func(t *testing.T) {
		type data struct {
			Hello string `json:"hello"`
		}
		d := data{"world"}
		err := Set("data", d)
		if err != nil {
			t.Fatal(err)
		}
		v, err := Get("data")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(v)

		var out data
		err = Scan("data", &out)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)
	})

	t.Run("max memory of one value", func(t *testing.T) {
		var value = "{\"ID\":3300579,\"ImportRecordID\":2105,\"MaterialVersionID\":1,\"MaterialID\":2,\"DeviceID\":1,\"Qualified\":true,\"BarCode\":\"##\",\"BarCodeStatus\":3,\"CreatedAt\":\"2020-08-18T00:35:44+08:00\",\"Attribute\":{},\"PointValues\":{\"FAI100\":9.794,\"FAI100-X\":7.681,\"FAI100-Y\":7.595,\"FAI16-1-B\":20.501,\"FAI16-1-F\":20.497,\"FAI16-2-B\":20.501,\"FAI16-2-F\":20.5,\"FAI16-3-B\":20.498,\"FAI16-3-F\":20.495,\"FAI17-1-B\":9.868,\"FAI17-1-F\":9.853,\"FAI17-2-B\":9.866,\"FAI17-2-F\":9.851,\"FAI17-3-B\":9.866,\"FAI17-3-F\":9.855,\"FAI18-1-B\":26.548,\"FAI18-1-F\":26.546,\"FAI18-2-B\":26.55,\"FAI18-2-F\":26.553,\"FAI18-3-B\":26.549,\"FAI18-3-F\":26.551,\"FAI23-1-B\":15.288,\"FAI23-1-F\":15.28,\"FAI23-2-B\":15.291,\"FAI23-2-F\":15.286,\"FAI23-3-B\":15.29,\"FAI23-3-F\":15.286,\"FAI24-1-B\":14.837,\"FAI24-1-F\":14.852,\"FAI24-2-B\":14.837,\"FAI24-2-F\":14.838,\"FAI24-3-B\":14.835,\"FAI24-3-F\":14.842,\"FAI35-1-B\":0.464,\"FAI35-1-F\":0.471,\"FAI35-2-B\":0.465,\"FAI35-2-F\":0.474,\"FAI35-3-B\":0.466,\"FAI35-3-F\":0.469,\"FAI38-1-B\":31.086,\"FAI38-1-F\":31.076,\"FAI38-2-B\":31.088,\"FAI38-2-F\":31.075,\"FAI38-3-B\":31.085,\"FAI38-3-F\":31.079,\"FAI81-1\":0.213,\"FAI81-2\":0.223,\"FAI81-3\":0.213,\"FAI81-4\":0.208,\"FAI95\":9.664,\"FAI95-X\":7.68,\"FAI95-Y\":23.019,\"FAI96\":0.044,\"FAI97\":0.041,\"FAI98\":8.619,\"FAI98-X\":21.035,\"FAI98-Y\":15.308,\"FAI99\":0.035}}"
		var sum = ""
		for i := 0; i < 5000; i++ {
			sum = sum + value
		}
		s1 := sum
		sum = sum + sum // 10000
		sum = sum + sum // 20000
		sum = sum + s1  // 25000
		fmt.Println(len(sum))
		fmt.Println(len("h"))

		var err error
		t1 := time.Now()
		err = Set("long_pds1", sum)
		t2 := util.DebugTime(t1, "set")
		if err != nil {
			t.Fatal(err)
		}
		_, err = Get("long_pds1")
		if err != nil {
			t.Fatal(err)
		}
		_ = util.DebugTime(t2, "get")
	})
}
