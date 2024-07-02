package dialect

import "reflect"

// Dialect 标准连接
type Dialect interface {
	DataTypeOf(typ reflect.Value) string            // 用于将 Go 语言的类型转换为该数据库的数据类型
	TableExistSQL(tableName string) (string, []any) // 返回某个表是否存在的 SQL 语句，参数是表名(table)
}

var dialectsMap = map[string]Dialect{}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect // 注册到全局
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
