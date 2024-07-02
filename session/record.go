package session

import (
	"GeeORM/clause"
	"errors"
	"log"
	"reflect"
)

// Insert 将每一个字段的值平铺
func (s *Session) Insert(values ...any) (int64, error) {
	recordValues := make([]any, 0)
	for _, value := range values {
		table := s.Model(value).RefTable() // 将结构体转化成表对象
		// 多次调用 clause.Set() 构造好每一个子句
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}
	s.clause.Set(clause.VALUES, recordValues...)
	// 调用一次 clause.Build() 按照传入的顺序构造出最终的 SQL 语句
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Find(values any) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values)) // 将切片转换为reflect.Value
	// destSlice.Type().Elem() 获取切片的单个元素的类型
	destType := destSlice.Type().Elem()
	log.Println("destType >>>>", destType)
	// New() 创建一个 destType 的实例, 作为 Model() 的入参，映射出表结构 RefTable()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	// 根据表结构，使用 clause 构造出 SELECT 语句
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	log.Println("组合的sql语句为: ", sql)
	// 查询出多条数据
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	for rows.Next() {
		dest := reflect.New(destType).Elem() // reflect.New(destType) 创建了一个指针，则 Elem() 方法返回该指针指向的变量
		var values []any
		for _, name := range table.FieldNames {
			// dest.FieldByName(name).Addr().Interface() 获取name字段的指针，指针地址不是重复的
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// Scan 将该行记录每一列的值依次赋值给 values 中的每一个字段，例如 err = rows.Scan(&id, &name, &age)
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

func (s *Session) Update(kv ...any) (int64, error) {
	// 平铺开来的键值对和 map 类型的键值对
	m, ok := kv[0].(map[string]any)
	if !ok {
		m = make(map[string]any)
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	log.Println(sql)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Delete() (int64, error) {
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var temp int64
	if err := row.Scan(&temp); err != nil {
		return 0, err
	}
	return temp, nil
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...any) *Session {
	var vars []any
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

func (s *Session) First(value any) error {
	// 获取value的反射值，并确保它是指针类型，解除指针引用
	dest := reflect.Indirect(reflect.ValueOf(value))
	// SliceOf 创建一个 User 类型的切片类型，New 创建切片的实例，Elem 获取反射值
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}
