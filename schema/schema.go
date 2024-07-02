package schema

import (
	"GeeORM/dialect"
	"go/ast"
	"reflect"
)

type Field struct {
	Name string // 字段名
	Type string
	Tag  string // 约束条件
}

type Schema struct {
	Model      any               // 对象
	Name       string            // 表名
	Fields     []*Field          // 字段
	FieldNames []string          // 包含所有的字段名(列名)
	fieldMap   map[string]*Field // 字段名和 Field 的映射关系
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

func (schema *Schema) RecordValues(dest any) []any {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []any
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

// Parse 将任意对象解析为Schema实例
func Parse(dest any, d dialect.Dialect) *Schema {
	// ValueOf 返回入参的值，返回的是一个对象的指针
	// Indirect 获取类型指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(), // 结构体的名称作为表名
		fieldMap: make(map[string]*Field),
	}
	// modelType.NumField 获取实例的字段的个数
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i) // 具体的字段实例
		// ast.IsExported 判断是否是大写字母开头的可导出的
		// p.Anonymous 判断字段是否是匿名字段
		if !p.Anonymous && ast.IsExported(p.Name) {
			// 当前字段结构体
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))), // 创建的新值的类型
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
