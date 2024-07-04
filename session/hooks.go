package session

import (
	"GeeORM/log"
	"reflect"
)

const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

//// 凡是实现以下接口的都将实现geeorm的钩子函数
//type IBeforeQuery interface {
//	BeforeQuery(s *Session) error
//}
//type IAfterQuery interface {
//	AfterQuery(s *Session) error
//}
//type IBeforeInsert interface {
//	BeforeInsert(s *Session) error
//}
//type IBeforeUpdate interface {
//	BeforeUpdate(s *Session) error
//}
//type IAfterUpdate interface {
//	AfterUpdate(s *Session) error
//}
//type IBeforeDelete interface {
//	BeforeDelete(s *Session) error
//}
//type IAfterDelete interface {
//	AfterDelete(s *Session) error
//}
//type IAfterInsert interface {
//	AfterInsert(s *Session) error
//}

func (s *Session) CallMethod(method string, value any) {
	// 获取名为 传入method 的方法
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	param := []reflect.Value{reflect.ValueOf(s)}
	// IsValid 判断是否有当前方法
	if fm.IsValid() {
		// Call 如果有，调用该方法并传参，接受一个 []reflect.Value 类型的切片作为参数列表
		// 返回一个 []reflect.Value 类型的切片
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	//fm := reflect.ValueOf(s.RefTable().Model)
	//if value != nil {
	//	fm = reflect.ValueOf(value)
	//}
	//call := func(i interface{}) {
	//	if err := reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{reflect.ValueOf(s)})[0].Interface(); err != nil {
	//		if e, ok := err.(error); ok && e != nil {
	//			log.Error(e)
	//		}
	//	}
	//}
	//switch method {
	//case BeforeQuery:
	//	if i, ok := fm.Interface().(IBeforeQuery); ok {
	//		call(i)
	//	}
	//case AfterQuery:
	//	if i, ok := fm.Interface().(IAfterQuery); ok {
	//		call(i)
	//	}
	//case BeforeInsert:
	//	if i, ok := fm.Interface().(IBeforeInsert); ok {
	//		call(i)
	//	}
	//}

	return
}
