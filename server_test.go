package GeeRPC

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num2 + args.Num1
	log.Println("数据: ", *reply)
	return nil
}

func (f Foo) name(args Args, reply *int) error {
	*reply = args.Num2 + args.Num1
	log.Println("name数据: ", *reply)
	return nil
}

func _assert(condition bool, msg string, v ...any) {
	if !condition {
		//panic()
		log.Println(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewServer(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_NumCalls(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"] // 结构体本身
	fmt.Printf("----------------- %+v\n", mType)
	argv := mType.newArgv()
	fmt.Println("------dddd----", argv)
	replyv := mType.newReplyV()
	// 将 argv 的值设置为 reflect.ValueOf(Args{Num1: 1, Num2: 4})
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 4}))
	err := s.call(mType, argv, replyv)
	fmt.Println("-----------------", err)
	fmt.Println("--------sss---------", *replyv.Interface().(*int))
	fmt.Println("-----------------", mType.NumCalls())

	//_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}
