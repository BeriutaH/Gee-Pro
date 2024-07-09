package geeorm

import (
	"GeeORM/session"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

//	func TestNewEngine(t *testing.T) {
//		engine := OpenDB(t)
//		defer engine.Close()
//	}
func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result any, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return nil, errors.New("自定义错误")
	})
	if err == nil || s.HasTable() {
		log.Println("回滚失败！")
	}
}

func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result any, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Beriuta", 17})
		return
	})
	u := &User{}
	_ = s.First(u)
	log.Printf("获取的数据>>> %+v", u)
	if err != nil || u.Name != "Tom" {
		log.Println("获取的名称出错")
	}
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

// go test -v -run TestEngine_Transaction
