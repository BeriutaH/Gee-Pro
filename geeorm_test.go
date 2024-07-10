package geeorm

import (
	"GeeORM/logger"
	"GeeORM/session"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"reflect"
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

func TestEngine_Migrate(t *testing.T) {
	logger.SetLevel(logger.InfoLevel)
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text PRIMARY KEY, XXX integer);").Exec()
	_, _ = s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	err := engine.Migrate(&User{})
	if err != nil {
		//log.Printf("迁移数据库错误信息: %v", err)
		logger.Error("迁移数据库错误信息: %v", err)
	}
	rows, _ := s.Raw("SELECT * FROM User").QueryRows()
	columns, _ := rows.Columns()
	logger.InfoF("获取到的字段为 %v", columns)
	if !reflect.DeepEqual(columns, []string{"Name", "Age"}) {
		log.Printf("迁移数据库失败，获取到的字段为: %v", columns)
	}

}

// go test -v -run TestEngine_Transaction
