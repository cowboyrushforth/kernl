package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type People struct {
  Kernl
}

func (c People) Show(guid string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  _, err := models.UserFromGuid(rc,guid)
  if err != nil {
    return c.NotFound("user not found")
  }
  x := []string{}
  return c.RenderJson(x)
}

