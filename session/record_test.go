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

func TestSession_Limit(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	err := s.Limit(1).Find(&users)
	if err != nil || len(users) != 1 {
		log.Println("查询一条信息失败")
	}
	log.Printf("用户列表: %+v\n", users)
}

func TestSession_Update(t *testing.T) {
	s := testRecordInit(t)
	affected, err := s.Where("Name = ?", "Tom").Update("Age", 30)
	if err != nil {
		log.Println(err)
	}
	u := &User{}
	_ = s.OrderBy("Age DESC").First(u)
	if affected != 1 || u.Age != 30 {
		log.Println("修改信息失败")
	}
	log.Printf("用户信息: %+v\n", u)
}

func TestSession_DeleteAndCount(t *testing.T) {
	s := testRecordInit(t)
	affected, _ := s.Where("Name = ?", "Tom").Delete()
	count, _ := s.Count()
	if affected != 1 || count != 1 {
		log.Println("删除信息失败")
	}
}

// go test -v ./session -run TestSession_Limit
