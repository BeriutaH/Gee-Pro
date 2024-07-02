package schema

import (
	"GeeORM/dialect"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{Name: "Jack", Age: 33}, TestDial)
	if schema.Name != "User" || len(schema.Fields) != 2 {
		t.Fatal("User结构体解析失败")
	}
	if schema.GetField("Name").Tag != "PRIMARY KEY" {
		t.Fatal("主键解析失败")
	}
}
