package utils

import "github.com/jinzhu/copier"

func Copy(toValue interface{}, fromValue interface{}) bool {
	err := copier.Copy(toValue, fromValue)
	if err != nil {
		return false
	}
	return true
}
