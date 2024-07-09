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
type TxFunc func(*session.Session) (any, error)

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

func (engine *Engine) Transaction(f TxFunc) (result any, err error) {
	// 一个新的session
	s := engine.NewSession()
	if err = s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		// 发生错误自动回滚
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			defer func() {
				if err != nil {
					_ = s.Rollback()
				}
			}()
			// 没有错误就提交
			err = s.Commit()
		}
	}()
	return f(s)
}
