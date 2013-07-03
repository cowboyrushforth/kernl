package controllers

import "github.com/robfig/revel"

type Home struct {
        Kernl
}

func (c Home) checkUser() revel.Result {
        if user := c.current_user(); user == nil {
                c.Flash.Error("Please log in first")
                return c.Redirect(Kernl.Index)
        }
        return nil
}

func (c Home) Index() revel.Result {
  current_user := c.current_user
  return c.Render(current_user)
}

