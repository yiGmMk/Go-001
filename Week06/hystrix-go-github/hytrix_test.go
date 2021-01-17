package main

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

func init() {

}

func TestHytrix(t *testing.T) {
	var nums sync.Map
	log.Println("start")
	var (
		SUCCESS string = "success"
		FAIL    string = "failure"
	)
	nums.Store(FAIL, 0)
	nums.Store(SUCCESS, 0)
	start := time.Now()
	for index := 0; index < 1000000; index++ {
		err := hystrix.Go("test ", func() error {
			suc, ok := nums.Load(SUCCESS)
			if !ok {
				log.Println("index:", index, "failed to load")
			}
			nums.Store(SUCCESS, suc.(int)+1)
			log.Println("index:", index, "success,size:", suc)
			return nil
		}, nil)
		if err != nil {
			fail, ok := nums.Load(FAIL)
			if !ok {
				log.Println("index:", index, "failed to load")
			}
			nums.Store(FAIL, fail.(int)+1)
			log.Println("index:", index, "failure,size:", fail)
		}
	}
	suc, ok := nums.Load(SUCCESS)
	fail, ok := nums.Load(FAIL)
	end := time.Now()
	log.Println("end,ok=", ok, "success:", suc, "failed:", fail, "time:", (end.Sub(start)).Seconds())
}
