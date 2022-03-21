package tool

import "github.com/duke-git/lancet/convertor"

func StringToInt(str string) int64 {
	i, err := convertor.ToInt(str)
	if err != nil {
		return 0
	}
	return i
}

func IntToString(i int64) string {
	return convertor.ToString(i)
}
