package controllers

import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "kernl/app/models"

func Init() {
  c, err := redis.Dial("tcp", ":6379")
  defer c.Close()
  if err != nil {
    panic(err)
  }
  defer c.Close()
}

type Kernl struct {
  *revel.Controller
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
    u, err := models.FetchUid(c.Session["uid"])
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
