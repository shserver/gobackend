package utility

import (
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

func SetCommand(c redis.Conn, key string, value interface{}) error {
	_, err := c.Do("SET", key, value)
	return err
}

// func Getommand(c redis.Conn, key string) (interface{}, error) {
// 	r, err := c.Do("GET", key)

// 	return err
// }

func Pipelining(c redis.Conn) {
	c.Send("SET", "test1", "111")
	c.Send("GET", "test2")
	c.Flush()
	r, _ := c.Receive() // reply from SET
	log.Print("r: ", r)
	r, _ = redis.Uint64(c.Receive()) // reply from GET
	log.Print("r: ", r)
}

func Pipelining2(c redis.Conn) {
	c.Send("MULTI")
	c.Send("SET", "test5", 444)
	c.Send("INCR", "test3")
	c.Send("GET", "asdf")
	r, _ := c.Do("EXEC")
	fmt.Println(r) // prints [1, 1]
}

func Subscribe(c redis.Conn) error {
	c.Send("SUBSCRIBE", "example")
	c.Flush()
	for {
		_, err := c.Receive()
		if err != nil {
			return err
		}
		// process pushed message
	}
}

func Publish(c redis.Conn) error {
	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe("example")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}
}

func Convert(c redis.Conn) {
	_, err := redis.Bool(c.Do("EXISTS", "foo"))
	if err != nil {
		// handle error return from c.Do or type conversion error.
	}
}

func MultiScan(c redis.Conn) {
	var value1 int
	var value2 string
	reply, err := redis.Values(c.Do("MGET", "key1", "key2"))
	if err != nil {
		// handle error
	}
	if _, err := redis.Scan(reply, &value1, &value2); err != nil {
		// handle error
	}
}

// Zpop
