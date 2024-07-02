package clause

import "strings"

type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]any
}

// Set 方法根据 Type 调用对应的 generator，生成该子句对应的 SQL 语句
func (c *Clause) Set(name Type, vars ...any) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]any)
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = vars
}

// Build 方法根据传入的 Type 的顺序，构造出最终的 SQL 语句
func (c *Clause) Build(orders ...Type) (string, []any) {
	var (
		sqlInfo []string
		vars    []any
	)
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqlInfo = append(sqlInfo, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqlInfo, " "), vars
}
