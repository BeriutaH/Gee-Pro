package geeorm

import (
	"GeeORM/dialect"
	"GeeORM/log"
	"GeeORM/session"
	"database/sql"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (engine *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// 发送 ping 以确保数据库连接处于活动状态
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	dial, ok := dialect.GetDialect(driver) // 获取 driver 对应的 dialect
	if !ok {
		log.Error("dialect %s Not Found", driver)
		return
	}
	engine = &Engine{db: db, dialect: dial}
	log.Info("链接数据库成功")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("数据库关闭失败", err)
	}
	log.Info("数据库关闭成功")
}

func (engine *Engine) NewSession() *session.Session {
	// 创建 Session 实例时，传递 dialect 给构造函数 New
	return session.New(engine.db, engine.dialect)
}
