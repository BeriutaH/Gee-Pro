package geeorm

import (
	"GeeORM/dialect"
	"GeeORM/logger"
	"GeeORM/session"
	"database/sql"
	"fmt"
	"strings"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}
type TxFunc func(*session.Session) (any, error)

func NewEngine(driver, source string) (engine *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		logger.Error(err)
		return
	}
	// 发送 ping 以确保数据库连接处于活动状态
	if err = db.Ping(); err != nil {
		logger.Error(err)
		return
	}

	dial, ok := dialect.GetDialect(driver) // 获取 driver 对应的 dialect
	if !ok {
		logger.Error("dialect %s Not Found", driver)
		return
	}
	engine = &Engine{db: db, dialect: dial}
	logger.Info("链接数据库成功")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		logger.Error("数据库关闭失败", err)
	}
	logger.Info("数据库关闭成功")
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

func (engine *Engine) Migrate(value any) error {
	_, err := engine.Transaction(func(s *session.Session) (result any, err error) {
		if !s.Model(value).HasTable() {
			logger.InfoF("表名: %s 不存在", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns) // 新表 - 旧表 = 新增字段
		delCols := difference(columns, table.FieldNames) // 旧表 - 新表 = 删除字段
		logger.InfoF("添加的字段: %v，删除的字段: %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			// ALTER 语句新增字段
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		// 创建新表并重命名的方式删除字段
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		// ALTER TABLE 修改数据库表 RENAME 重命名表
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}

// difference 获取a跟b中，不同的值
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}
