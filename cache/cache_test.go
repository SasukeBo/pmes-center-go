package cache

import (
	"fmt"
	"testing"
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
}
