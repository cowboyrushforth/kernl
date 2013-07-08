package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

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
    u, err := models.UserFromUid(c.Session["uid"])
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

func (c Kernl) checkUser() revel.Result {
  revel.INFO.Println("checking user")
  if user := c.current_user(); user == nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(Kernl.Index)
  }
  return nil
}
