package session

import (
	"log"
	"testing"
)

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 23}
	user3 = &User{"Jack", 48}
)

func testRecordInit(t *testing.T) *Session {
	t.Helper()
	s := NewSession().Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		log.Println(err1)
		log.Println(err2)
		log.Println(err3)
		log.Println("初始化记录错误")
	}
	return s
}

func TestSession_Insert(t *testing.T) {
	s := testRecordInit(t)
	affected, err := s.Insert(user3)
	if err != nil || affected != 1 {
		log.Println("创建记录失败")
	}
}

func TestSession_Find(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	if err := s.Find(&users); err != nil {
		log.Println("查询记录失败")
	}
	log.Println(users)
}
