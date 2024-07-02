package clause

import (
	"fmt"
	"testing"
)

func testSelect(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Age ASC")
	sqlInfo, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	fmt.Printf("sql 语句: %s\n vars信息: %v", sqlInfo, vars)
}

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
}
