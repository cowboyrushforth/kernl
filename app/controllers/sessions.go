package controllers

import "github.com/robfig/revel"
import "code.google.com/p/go.crypto/bcrypt"
import "kernl/app/models"

type Sessions struct {
  Kernl
}

func (c Sessions) New() revel.Result {
  return c.Render()
}

func (c Sessions) Create(slug string, password string) revel.Result {
  // TODO Validations
  rc := GetRedisConn()
  defer rc.Close()
  u, err := models.UserFromSlug(rc, slug)
  if err == nil {
    errb := bcrypt.CompareHashAndPassword(u.PwdHash, []byte(password))
    if errb == nil {
      c.Session["uid"] = "user:"+slug
      c.Flash.Success("Welcome, " + u.String())
      return c.Redirect(Home.Index)
    } 
  }
  c.Flash.Error("Login Failed")
  return c.Redirect(Sessions.New)
}

func (c Sessions) Destroy() revel.Result {
  c.Session["uid"] = ""
  return c.Redirect(Kernl.Index)
}
