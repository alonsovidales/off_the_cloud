package controllers

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

// In order to test the concurrenty we re going to do a benckmark test
func BenchmarkInsertShowAndDelete(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ds := GetDataStorageController()

	for i := 0; i < 3; i++ {
		b.Run(fmt.Sprintf("Stores: %d", i), func(b *testing.B) {
			for i := 0; i <= b.N; i++ {
				repo := fmt.Sprintf("repo_%d", i%100)
				val := []byte(fmt.Sprintf("val_%d", i))

				code, result := ds.New(repo, val)
				if code != 201 {
					b.Errorf("the expected code was 201 but %d was returned", code)
				}

				mapResult := result.(map[string]interface{})
				oid := mapResult["oid"].(string)

				code, result = ds.Show(repo, oid)
				if code != 200 {
					b.Errorf("the expected code was 200 but %d was returned", code)
				}

				if !reflect.DeepEqual(val, result) {
					b.Errorf("stored value: %v, returned: %v", val, result)
				}

				code, _ = ds.Destroy(repo, oid)
				if code != 200 {
					b.Errorf("the expected code was 200 but %d was returned", code)
				}
				code, _ = ds.Show(repo, oid)
				if code != 404 {
					b.Errorf("the expected code after a destroy is 404 but %d was returned", code)
				}
			}
		})
	}
}
