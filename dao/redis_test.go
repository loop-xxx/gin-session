package dao

import (
	"fmt"
	"testing"
	"time"
)

func TestRedisClient_Push(t *testing.T) {
	if r, err := DefaultRedis("192.168.20.130:6379", "toor", 0); err == nil{
		defer r.Done()
		if err := r.Push("gs:uuid", "loop", map[string]string{"name": "loop", "age":"25"}, time.Second*30); err != nil{
			t.Error(err)
		}
	}else{
		t.Error(err)
	}
}

func TestRedisClient_Pull(t *testing.T) {
	if r, err := DefaultRedis("192.168.20.130:6379", "toor", 0); err == nil{
		defer r.Done()

		if magic, json, err := r.Pull("gs:uuid"); err == nil{
			fmt.Println(magic, json)
		}else {
			t.Error(err)
		}
	}else{
		t.Error(err)
	}
}

func TestRedisClient_Check(t *testing.T) {
	if r, err := DefaultRedis("192.168.20.130:6379", "toor", 0); err == nil{
		defer r.Done()
		time.Sleep(time.Second*6)
		if status, err := r.Check("gs:uuid", "loop", time.Second *30); err == nil{
			fmt.Println(status)
		}else{
			t.Error(err)
		}
	}else{
		t.Error(err)
	}
}

func TestRedisClient_Peek(t *testing.T) {
	if r, err := DefaultRedis("192.168.20.130:6379", "toor", 0); err == nil{
		defer r.Done()
		time.Sleep(time.Second*6)
		if ok, err := r.Peek("gs:uuid", "loop"); err == nil{
			fmt.Println(ok)
		}else{
			t.Error(err)
		}
	}else{
		t.Error(err)
	}
}