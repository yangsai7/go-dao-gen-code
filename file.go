package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	FileAlreadyExistErr = errors.New("file already exist")
)

func GetTableFile(fileName string) (f *os.File, err error) {
	// trim `_` character from fileName
	fileName = strings.Replace(fileName, "_", "", -1)
	if fileName == "" {
		return f, errors.New("error: fileName can not be empty")
	}
	fileName += ".go"
	filePath := filepath.Join(opDir, fileName)
	if _, err = os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return os.Create(filePath)
		}
		return
	}
	fmt.Printf("file %s already exist, do you want to overwrite it? [y/n]: ", fileName)
	var op string
	fmt.Scanf("%s", &op)
	if strings.ToLower(op) == "y" {
		return os.Create(filePath)
	}
	return f, FileAlreadyExistErr
}

func GetTableCondsFile(fileName string) (f *os.File, err error) {
	// trim `_` character from fileName
	fileName = strings.Replace(fileName, "_", "", -1)
	if fileName == "" {
		return f, errors.New("error: fileName can not be empty")
	}
	fileName += "conds.go"
	filePath := filepath.Join(opDir, fileName)
	if _, err = os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return os.Create(filePath)
		}
		return
	}
	fmt.Printf("file %s already exist, do you want to overwrite it? [y/n]: ", fileName)
	var op string
	fmt.Scanf("%s", &op)
	if strings.ToLower(op) == "y" {
		return os.Create(filePath)
	}
	return f, FileAlreadyExistErr
}

func GetInitDaoFile() (f *os.File, err error) {
	fileName := "dao.go"
	filePath := filepath.Join(opDir, fileName)
	if _, err = os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return os.Create(filePath)
		}
		return
	}
	fmt.Printf("file %s already exist, do you want to overwrite it? [y/n]: ", fileName)
	var op string
	fmt.Scanf("%s", &op)
	if strings.ToLower(op) == "y" {
		return os.Create(filePath)
	}
	return f, FileAlreadyExistErr
}
