package main

import "time"

type Count struct {
	Cur      time.Time     `json:"cur" description:"当前时间"`
	TimeSpan time.Duration `json:"time_span" description:"持续计数时间,窗口时间"`
	Bucket   int           `json:"bucket" description:""`
	Total    int64         `json:"total" description:"总数"`
	Fail     int64         `json:"fail" description:"失败数"`
	Success  int64         `json:"success" description:""`
	Timeout  int64         `json:"timeout" description:""`
	Reject   int64         `json:"reject" description:""`
}

func (c *Count) addSuccess() int64 {
	return 0
}

func (c *Count) addFail() int64 {
	return 0
}

func (c *Count) getTotal() int64 {
	return 0
}

func (c *Count) CanPass() bool {
	return true
}

func main() {

}
