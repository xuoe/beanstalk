package beanstalk_test

import (
	"fmt"
	"time"

	"github.com/xuoe/beanstalk"
)

func Example() {
	c, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
	if err != nil {
		panic(err)
	}
	c.Put([]byte("hello"), 1, 0, 120*time.Second)
	id, body, err := c.ReserveWithTimeout(5 * time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Println("job", id)
	fmt.Println(string(body))
}

func ExampleTubeSet_ReserveWithTimeout() {
	c, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
	if err != nil {
		panic(err)
	}
	id, body, err := c.ReserveWithTimeout(10 * time.Hour)
	if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
		fmt.Println("timed out")
		return
	} else if err != nil {
		panic(err)
	}
	fmt.Println("job", id)
	fmt.Println(string(body))
}
