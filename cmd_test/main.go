package main

import (
	geeorm "GeeORM"
	_ "github.com/mattn/go-sqlite3" // 导入时会注册 sqlite3 的驱动
	"log"
)

func main() {
	//db, err := sql.Open("sqlite3", "gee.db") // 返回一个 sql.DB 实例的指针
	//if err != nil {
	//	log.Println("链接数据库错误: ", err)
	//}
	////defer db.Close()
	//defer func() { _ = db.Close() }()
	//_, err = db.Exec("DROP TABLE IF EXISTS User;")
	//if err != nil {
	//	log.Println("删除表格错误: ", err)
	//	return
	//}
	//
	//_, err = db.Exec("CREATE TABLE User(Name text);")
	//if err != nil {
	//	log.Println("创建表格错误: ", err)
	//	return
	//}
	//
	////db.Exec("DROP TABLE IF EXISTS User;")
	////db.Exec("CREATE TABLE User(Name text);")
	//result, err := db.Exec("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam")
	//if err == nil {
	//	affected, _ := result.RowsAffected()
	//	log.Println(affected)
	//}
	//// Query() 和 QueryRow()，前者可以返回多条记录，后者只返回一条记录
	//row := db.QueryRow("SELECT Name FROM User LIMIT 1")
	//var name string
	//if err = row.Scan(&name); err == nil {
	//	log.Println(name)
	//}
	engine, _ := geeorm.NewEngine("sqlite3", "gee.db")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	log.Printf("执行成功, %d 条受影响\n", count)

}
