package csv

import (
	"reflect"
	"testing"
	"time"
)

func TestWeek(t *testing.T) {
	var d struct {
		List  []int
		Map   map[int]int
		Week1 Week
		Week2 Week
		Date  time.Time
	}

	l := "1,2,3"
	m := "1:1,2:2,3:3"
	w1 := "12:00:00"
	w2 := "5-03:00:00"
	date := "2019-01-01 11:11:11"
	v := reflect.ValueOf(&d)
	storeValue(l, v.Elem().Field(0))
	storeValue(m, v.Elem().Field(1))
	storeValue(w1, v.Elem().Field(2))
	storeValue(w2, v.Elem().Field(3))
	storeValue(date, v.Elem().Field(4))
	t.Log(d)
}
