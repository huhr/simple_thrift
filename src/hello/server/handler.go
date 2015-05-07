package server

import (
	"gen/hello"
)

// 接口的实现类,通常业务逻辑较为复杂时
// 我们会多分出去一层
type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetUser(uid int32) (r *hello.User, err error) {
	r = hello.NewUser()
	r.Firstname = "hu"
	r.Lastname = "haoran"
	r.Id = 1
	return
}

func (h *Handler) AddUser(user *hello.User) (r bool, err error) {
	return
}
