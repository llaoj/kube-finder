package osutil

import (
	"os"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func FileName(path string) string {
	s, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return s.Name()
}

//func IsFile(path string) bool {
//	return !IsDir(path)
//}
