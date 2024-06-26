package geeorm

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("gee数据库连接失败", err)
	}
	return engine
}

func TestNewEngine(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
}

// go test -v -run TestNewEngine
