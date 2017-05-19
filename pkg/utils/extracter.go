package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Extracter struct {
	m   map[string]interface{}
	v   reflect.Value
	err error
}

func NewExtracter(m map[string]interface{}) *Extracter {
	return &Extracter{m: m}
}

func (e *Extracter) GetError() error {
	return e.err
}

func (e *Extracter) Get(path string) (extracter *Extracter) {
	extracter = e
	if len(path) == 0 {
		e.err = errors.New("param 'path' is empty string")
		return
	}

	valuemap := e.m
	components := strings.Split(path, ".")
	for i := 0; i < len(components)-1; i++ {
		v, ok := valuemap[components[i]]
		if !ok {
			e.err = fmt.Errorf("there is no field named '%s'", components[i])
			return
		}
		m, assertok := v.(map[string]interface{})
		if !assertok {
			e.err = fmt.Errorf("field '%d' is not object", components[i])
			return
		}
		valuemap = m
	}

	v, ok := valuemap[components[len(components)-1]]
	e.v = reflect.ValueOf(v)
	if !ok {
		e.err = fmt.Errorf("no entry named '%s'\n", components[len(components)-1])
	}
	return
}

func (e *Extracter) CheckFloat64() (extracter *Extracter) {
	extracter = e
	if e.err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			e.err = fmt.Errorf("check float64 failed: %s", e.v.Kind().String())
		}
	}()
	e.v = e.v.Convert(reflect.TypeOf(float64(0)))
	return
}

func (e *Extracter) CheckUint32() (extracter *Extracter) {
	extracter = e
	if e.err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			e.err = fmt.Errorf("check uint32 failed: %s", e.v.Kind().String())
		}
	}()
	e.v = e.v.Convert(reflect.TypeOf(uint32(0)))
	return
}

func (e *Extracter) CheckInt() (extracter *Extracter) {
	extracter = e
	if e.err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			e.err = fmt.Errorf("check int failed: %s", e.v.Kind().String())
		}
	}()
	e.v = e.v.Convert(reflect.TypeOf(int(0)))
	return
}

func (e *Extracter) ToFloat64() (float64, error) {
	if e.err != nil {
		return 0, e.err
	}
	return e.v.Float(), nil
}

func (e *Extracter) ToUint32() (uint32, error) {
	if e.err != nil {
		return 0, e.err
	}
	return uint32(e.v.Uint()), nil
}

func (e *Extracter) ToInt() (int, error) {
	if e.err != nil {
		return 0, e.err
	}
	return int(e.v.Int()), nil
}
