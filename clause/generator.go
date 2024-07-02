package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...any) (string, []any)

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
}

func getBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

func _orderBy(values ...any) (string, []any) {
	// 按照指定字段排序
	return fmt.Sprintf("ORDER BY %s", values[0]), []any{}
}

func _where(values ...any) (string, []any) {
	// WHERE $desc  根据条件查询
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _limit(values ...any) (string, []any) {
	// LIMIT $num 用于强制 SELECT 语句返回指定的记录数
	return "LIMIT ?", values
}

func _select(values ...any) (string, []any) {
	// SELECT $fields FROM $tableName 查询指定字段信息
	tableName := values[0]
	fiields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fiields, tableName), []any{}
}

func _values(values ...any) (string, []any) {
	// VALUES ($v1), ($v2), ... 指定插入的字段对应的值
	var bindStr string
	var sql strings.Builder
	var vars []any
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]any)
		if bindStr == "" {
			bindStr = getBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars

}

func _insert(values ...any) (string, []any) {
	// INSERT INTO $tableName ($fields)  插入一条数据
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []any{}
}
