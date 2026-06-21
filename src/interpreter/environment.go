package interpreter

import (
	tunaparser "tunascript/src/parser"
	"fmt"
)

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables:	map[string]RuntimeValue{},
		constants:	map[string]bool{},
		declaredTypes:	map[string]tunaparser.AstType{},
		parent:		parent,
	}
}

func (e *Environment) Update(name string, val RuntimeValue) (RuntimeValue, error) {
	if _, ok := e.variables[name]; ok {
		if e.constants[name] {
			return RuntimeValue{}, fmt.Errorf("cannot reassign constant '%s'", name)
		}
		e.variables[name] = val
		return val, nil
	}
	if e.parent != nil {
		return e.parent.Update(name, val)
	}
	return RuntimeValue{}, fmt.Errorf("cannot assign to undefined variable '%s'", name)
}

func (e *Environment) MustUpdate(name string, val RuntimeValue) RuntimeValue {
	v, err := e.Update(name, val)
	if err != nil {
		panic(TunaError(err.Error()))
	}
	return v
}

func (e *Environment) Set(name string, val RuntimeValue) RuntimeValue {
	e.variables[name] = val
	return val
}

func (e *Environment) SetTyped(name string, val RuntimeValue, t tunaparser.AstType) RuntimeValue {
	e.variables[name] = val
	if t != nil {
		e.declaredTypes[name] = t
	}
	return val
}

func (e *Environment) SetConst(name string, val RuntimeValue) RuntimeValue {
	e.variables[name] = val
	e.constants[name] = true
	return val
}

func (e *Environment) Get(name string) (RuntimeValue, bool) {
	if val, ok := e.variables[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return RuntimeValue{}, false
}

func (e *Environment) MustGet(name string) RuntimeValue {
	v, ok := e.Get(name)
	if !ok {
		panic(TunaError(fmt.Sprintf("undefined variable '%s'", name)))
	}
	return v
}

func (e *Environment) GetDeclaredType(name string) tunaparser.AstType {
	if t, ok := e.declaredTypes[name]; ok {
		return t
	}
	if e.parent != nil {
		return e.parent.GetDeclaredType(name)
	}
	return nil
}
