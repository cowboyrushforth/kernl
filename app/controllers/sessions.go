package controllers

import "github.com/robfig/revel"
import "code.google.com/p/go.crypto/bcrypt"
import "kernl/app/models"
import "net/url"

type Sessions struct {
  Kernl
}

func (c Sessions) New(return_to string) revel.Result {
  return c.Render(return_to)
}

func (c Sessions) Create(slug string, password string, return_to string) revel.Result {
  u, err := models.UserFromSlug(slug)
  if err == nil {
    errb := bcrypt.CompareHashAndPassword(u.PwdHash, []byte(password))
    if errb == nil {
      c.Session["uid"] = "user:"+slug
      c.Flash.Success("Welcome, " + u.String())
      if return_to == "" {
        return c.Redirect(Home.Index)
      } else {
        return c.Redirect(return_to)
      }
    } 
  }
  c.Flash.Error("Login Failed")
  if return_to != "" {
    return c.Redirect("/login?return_to="+url.QueryEscape(return_to))
  }
  return c.Redirect(Sessions.New)
}

func (c Sessions) Destroy() revel.Result {
  c.Session["uid"] = ""
  return c.Redirect(Kernl.Index)
}
