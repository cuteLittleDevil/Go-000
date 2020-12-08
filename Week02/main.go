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
	sql := ""
	err := errors.New("sql 错误")
	return fmt.Errorf("dao sql is: %s custom err is: %w original err is: %s", sql, ERR_SQL, err)
}
