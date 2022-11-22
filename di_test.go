package di

import (
	"log"
	"testing"
)

type Person interface {
	SayHello()
}

type P1 struct {
	Name string
}

func (p P1) SayHello() {
	log.Printf("%s say hello", p.Name)
}

type Home struct {
	XM Person `di:"api"`
	Li *P1    `di:"val"`
}

func TestNewContainer(t *testing.T) {
	ioc := NewContainer()

	p1 := P1{Name: "xiao ming"}
	li := P1{Name: "Li"}
	home := Home{}

	ioc.RegisterApi((*Person)(nil), p1).
		RegisterVal(li).
		ResolveField(&home)

	ioc.Init()

	home.XM.SayHello()
	li.SayHello()
}
