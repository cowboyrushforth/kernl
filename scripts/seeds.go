package main

import "fmt"
import "github.com/garyburd/redigo/redis"

func main() {
  c, err := redis.Dial("tcp", ":6379")
  if err == nil {
    fmt.Println("Hello, 世界", c)
  }
}
