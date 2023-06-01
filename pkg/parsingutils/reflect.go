package parsingutils

import "reflect"

func GetDataValue(v interface{}) reflect.Value {
	dat := reflect.ValueOf(v)
	for dat.Kind() == reflect.Ptr {
		dat = dat.Elem()
	}
	return dat
}
