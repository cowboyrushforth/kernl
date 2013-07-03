package controllers

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "kernl/app/models"
import "time"

var redisPool *redis.Pool

func Init() {
  redisPool = redis.NewPool(func() (redis.Conn, error) {
    c, err := redis.Dial("tcp", ":6379")
    if err != nil {
      return nil, err
    }
    return c, err
  }, 3)
}

type Kernl struct {
  *revel.Controller
}

func GetRedisConn() redis.Conn {
        rc := redisPool.Get()
        for i := 0; i < 5; i++ {
                if err := rc.Err(); err != nil {
                        time.Sleep(10 * time.Millisecond)
                        rc = redisPool.Get()
                } else {
                        break
                }
        }
        return rc
}

func (c Kernl) Index() revel.Result {
  if c.current_user() != nil {
    return c.Redirect(Home.Index)
  }
  return c.Render()
}

func (c Kernl) current_user() *(models.User) {
  if c.RenderArgs["user"] != nil {
    return c.RenderArgs["user"].(*models.User)
  }
  if c.Session["uid"] != "" {
    revel.INFO.Println("auth check uid", c.Session["uid"])
    // see if uid is valid
    rc := GetRedisConn()
    defer rc.Close()
    u, err := models.FetchUid(rc, c.Session["uid"])
    if err != nil {
      revel.INFO.Println("\tauth BAD!", err)
      c.Session["uid"] = ""
    } else {
      revel.INFO.Println("\tauth good", u.Email)
      c.RenderArgs["user"] = u
      return u 
    }
  }
  return nil
}
