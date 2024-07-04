package session

import (
	"log"
	"testing"
)

type Account struct {
	ID       int `geeorm:"PRIMARY KEY"`
	Password string
}

func (a *Account) BeforeInsert(s *Session) error {
	log.Println("插入之前: ", a)
	a.ID += 1000
	return nil
}

func (a *Account) AfterQuery(s *Session) error {
	log.Println("查新之后: ", a)
	a.Password = "******"
	return nil
}

func TestSession_CallMethod(t *testing.T) {
	s := NewSession().Model(&Account{})
	err := s.DropTable()
	if err != nil {
		log.Println("删除表失败: ", err)
	}
	err = s.CreateTable()
	if err != nil {
		log.Println("创建表失败: ", err)
	}
	_, err = s.Insert(&Account{1, "123456"}, &Account{2, "789012"})
	if err != nil {
		log.Println("插入信息失败: ", err)
	}
	u := Account{}
	err = s.First(&u)
	log.Printf("%+v", s.RefTable().Model)
	log.Println("用户账号信息: ", u.ID, u.Password)

}
