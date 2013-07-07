package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type People struct {
  KernlAuthed
}

func (c People) Show(slug string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  _, err := models.UserFromSlug(rc,slug)
  if err != nil {
    return c.NotFound("user not found")
  }
  x := []string{}
  return c.RenderJson(x)
}

func (c People) ShowRemote(guid string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  person, err := models.PersonFromGuid(rc,guid)
  if err != nil {
    return c.NotFound("person not found")
  }
  return c.Render(person)
}
