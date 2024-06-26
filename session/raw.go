package session

import (
	"GeeORM/dialect"
	"GeeORM/log"
	"GeeORM/schema"
	"database/sql"
	"strings"
)

type Session struct {
	db       *sql.DB // 用 sql.Open() 方法连接数据库成功之后返回的指针
	dialect  dialect.Dialect
	refTable *schema.Schema
	// 用户调用 Raw() 方法即可改变这两个变量的值
	sql     strings.Builder // 拼接 SQL 语句 使用 strings.Builder 避免每次修改其实都要重新申请一个内存空间
	sqlVars []any           // SQL 语句中占位符的对应值
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// Clear 清空 (s *Session).sql 和 (s *Session).sqlVars 两个变量
func (s *Session) Clear() {
	s.sql.Reset() // 重置 strings.Builder 类型为空
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.db
}

func (s *Session) Raw(sql string, values ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec 执行原始 sql
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow 从数据库获取一条记录
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 从数据库获取记录列表
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
