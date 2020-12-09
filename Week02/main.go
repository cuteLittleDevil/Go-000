package main

import (
	"errors"
	"fmt"
)

var ERR_SQL = errors.New("sql")

func main() {
	if errors.Is(dao(), ERR_SQL) {
		fmt.Println(dao())
	}
}

func dao() error {
	sql := "select * from week"
	err := errors.New("sql 错误")
	return NewDaoErr(ERR_SQL, err, fmt.Sprintf("sql is %s", sql))
}

func NewDaoErr(custom, original error, info string) error {
	return fmt.Errorf("[info is %s] [custom err is %w] [original err is %s]", info, custom, original)
}
