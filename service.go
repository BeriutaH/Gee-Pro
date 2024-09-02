package GeeRPC

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method // 方法本身
	ArgType   reflect.Type   // 第一个参数的类型
	ReplyType reflect.Type   // 第二个参数的类型
	numCalls  uint64         // 后续统计方法调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	//  检查 ArgType 是否是指针类型
	if m.ArgType.Kind() == reflect.Ptr {
		// 获取指针所指向的元素类型，并创建一个新的该类型的值
		argv = reflect.New(m.ArgType.Elem())
	} else {
		// 创建一个新的 ArgType 类型的值，并调用 Elem() 方法获取指针指向的具体值
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyV() reflect.Value {
	replyV := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyV.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyV.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	default:
		log.Println("未处理的默认情况")
	}
	return replyV
}

type service struct {
	name   string                 // 映射的结构体的名称
	typ    reflect.Type           // 结构体的类型
	rcvr   reflect.Value          // 结构体的实例本身
	method map[string]*methodType // 存储映射的结构体的所有符合条件的方法
}

func (s *service) registerMethods() {
	/*
		两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身）
		返回值有且只有 1 个，类型为 error
	*/
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		// NumIn 获取入参个数， NumOut 获取输出参数个数
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// 检查方法或函数的第一个返回值类型是否为 error
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc 服务器：注册 %s.%s\n", s.name, method.Name)
	}
}

func (s *service) call(m *methodType, argv, replyV reflect.Value) error {
	log.Println("调用方法次数: ", m.numCalls)
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func // 获取方法
	// Call 调用方法，3个参数，第一个是实例本身
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyV})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func newService(rcvr any) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc 服务器: %s 不是有效的服务名称", s.name)
	}
	s.registerMethods()
	return s
}
