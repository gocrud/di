package di

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Container interface {
	// RegisterApi 注入接口
	//
	// 示例：di.NewContainer().RegisterApi((*di.IUser)(nil), &di.User{})
	RegisterApi(inter any, implement any) Container
	// RegisterVal 注入值
	//
	// 示例：di.NewContainer().RegisterValue(&User{})
	RegisterVal(value any) Container
	// RegisterKey 注入token值
	//
	// 示例：di.NewContainer().RegisterKey("token",&User{})
	RegisterKey(token string, value any) Container
	// ResolveField 字段依赖
	//
	// 示例：di.NewContainer().ResolveField(&user{})
	ResolveField(values ...any) Container
	// ResolveVal 值依赖
	//
	// 示例：di.NewContainer().ResolveValue(&user{})
	ResolveVal(values ...any) Container
	// Init 容器初始化函数
	Init() error
	// AfterInit 容器初始化后执行函数
	AfterInit(...func() error)
}

type container struct {
	registerStore sync.Map
	mux           sync.RWMutex
	resolveStore  map[string][]any
	initAfterFunc []func() error
}

func (c *container) Init() error {
	for k, v := range c.resolveStore {
		switch k {
		case "filed":
			for _, vv := range v {
				ref := reflect.ValueOf(vv)
				typeRef := reflect.TypeOf(vv)
				for i := 0; i < ref.Elem().NumField(); i++ {
					depFiled := ref.Elem().Field(i)
					var fieldTag = getTag(typeRef.Elem().Field(i).Tag.Get("di"))
					switch fieldTag.key {
					case "api":
						c.resolveApi(depFiled)
					case "val":
						c.resolveVal(depFiled)
					case "key":
						c.resolveKey(fieldTag.value, depFiled)
					}
				}
			}
		case "value":
			for _, vv := range v {
				vRef := reflect.ValueOf(vv)
				c.resolveVal(vRef)
			}
		}

	}

	// 执行初始化后函数
	for _, v := range c.initAfterFunc {
		if err := v(); err != nil {
			return err
		}
	}
	c.resolveStore = nil
	return nil
}

type tag struct {
	key   string
	value string
}

func getTag(s string) tag {
	tags := strings.Split(s, ":")
	var t tag
	switch tags[0] {
	case "val", "api":
		t.key = tags[0]
	case "key":
		if len(tags) == 2 {
			t.key = tags[0]
			t.value = tags[1]
		}
	}
	return t
}

func (c *container) AfterInit(args ...func() error) {
	c.initAfterFunc = append(c.initAfterFunc, args...)
}

func (c *container) ResolveField(values ...any) Container {
	c.mux.Lock()
	c.resolveStore["filed"] = append(c.resolveStore["filed"], values...)
	c.mux.Unlock()
	return c
}

func (c *container) ResolveVal(values ...any) Container {
	c.mux.Lock()
	c.resolveStore["value"] = append(c.resolveStore["value"], values...)
	c.mux.Unlock()
	return c
}

func (c *container) resolveVal(ref reflect.Value) {
	kindStr := getValueKindStr(ref)
	val, _ := c.registerStore.Load(kindStr)
	valRef := reflect.ValueOf(val)
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			f := reflect.New(ref.Type().Elem())
			ref.Set(f)
		}
		ref = ref.Elem()
	}
	if valRef.Kind() == reflect.Ptr {
		ref.Set(valRef.Elem())
	} else {
		ref.Set(valRef)
	}
}

func (c *container) resolveApi(ref reflect.Value) {
	filedKind := getValueKindStr(ref)
	val, ok := c.registerStore.Load(filedKind)
	if !ok {
		return
	}
	valRef := reflect.ValueOf(val)
	ref.Set(valRef)
}

func (c *container) resolveKey(token string, ref reflect.Value) {
	val, ok := c.registerStore.Load(token)
	if !ok {
		return
	}
	valRef := reflect.ValueOf(val)
	ref.Set(valRef)
}

func (c *container) RegisterApi(inter any, implement any) Container {
	var interRef = reflect.TypeOf(inter)
	// 判断是否是接口
	if interRef.Elem().Kind() != reflect.Interface {
		panic("inter 必须是接口类型")
	}
	// 判断是否实现了接口
	var impleRef = reflect.TypeOf(implement)
	if !impleRef.Implements(interRef.Elem()) {
		panic(fmt.Sprintf("%s必须是%v的实现类", impleRef.Elem().String(), interRef.Elem()))
	}
	c.registerStore.Store(getTypeKindStr(interRef), implement)
	return c
}

func (c *container) RegisterKey(token string, value any) Container {
	c.registerStore.Store(token, value)
	return c
}

func (c *container) RegisterVal(value any) Container {
	ref := reflect.ValueOf(value)
	kindStr := getValueKindStr(ref)
	c.registerStore.Store(kindStr, value)
	return c
}

type InvokeOptions func(c *container)

func NewContainer() Container {
	return &container{
		registerStore: sync.Map{},
		resolveStore:  make(map[string][]any),
	}
}
