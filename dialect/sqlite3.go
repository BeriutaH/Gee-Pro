package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type sqlite3 struct{}

var _ Dialect = (*sqlite3)(nil)

// init 包在第一次加载时，会将 sqlite3 的 dialect 自动注册到全局
func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

func (s *sqlite3) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float64, reflect.Float32:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	default:
		panic("unhandled default case")
	}
	panic(fmt.Sprintf("无效的 SQL 类型 %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (s *sqlite3) TableExistSQL(tableName string) (string, []any) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}
