package main

import "fmt"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "kernl/app/models"
import "code.google.com/p/go.crypto/bcrypt"

func main() {
  revel.append(ConfPaths, '/Users/aa/scott/src/kernl/conf')
  config, err := revel.LoadConfig("app.conf")
  if err != nil {
    fmt.Println("Hello", err)
    fmt.Println("Hello", config)
    return
  }

  c, err := redis.Dial("tcp", "127.0.0.1:6379")
  if err == nil {
    user := models.User{}
    user.Slug = "andrew"
    user.Email = "andrew@aa.com"

    user.PwdHash, _ = bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
    fmt.Println("Hello, 世界", user, c)
    err := user.Insert(c)
    if err == true {
      fmt.Println("Hello, 世界", user)
    }
  }
}
