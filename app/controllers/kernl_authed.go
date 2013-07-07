package controllers

import "github.com/robfig/revel"

type KernlAuthed struct {
  Kernl
}

func init() {
  revel.InterceptMethod(KernlAuthed.checkUser, revel.BEFORE)
}

func (c KernlAuthed) checkUser() revel.Result {
  revel.INFO.Println("checking user")
  if user := c.current_user(); user == nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(Kernl.Index)
  }
  return nil
}
