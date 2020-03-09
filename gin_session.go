package session

import (
	"encoding/json"
	"errors"
)

type GinSession struct{
	flag bool
	core map[string] string
}

func New(data map[string]string)(g* GinSession){
	g = &GinSession{
		flag: false,
		core: data,
	}
	return
}

func (g *GinSession)Dump() (data map[string]string){
	data = g.core
	return
}

func (g * GinSession) Check()(flag bool){
	flag = g.flag
	return
}

func (g *GinSession)Get(key string)(value string, ok bool){
	value, ok = g.core[key]
	return
}

func (g *GinSession)Set(key string, value string){
	g.flag = true
	g.core[key] = value
}

func (g *GinSession)SetStruct(key string, value interface{})(err error){
	if bytes, marshalErr := json.Marshal(value); marshalErr == nil{
		g.Set(key, string(bytes))
	}else{
		err = marshalErr
	}
	return
}

func (g *GinSession)GetStruct(key string, pointer interface{})(err error){
	if value, ok:= g.core[key]; ok{
		err = json.Unmarshal([]byte(value), pointer)
	}else{
		err = errors.New("key does not exist")
	}
	return
}