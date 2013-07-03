package controllers

import "github.com/robfig/revel"

type Connections struct {
  Kernl
}

func (c Connections) checkUser() revel.Result {
  if user := c.current_user(); user == nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(Kernl.Index)
  }
  return nil
}

func (c Connections) Index() revel.Result {
  return c.Render()
}

func (c Connections) New() revel.Result {
  return c.Render()
}

func (c Connections) Verify() revel.Result {
  return c.Render()
}

func (c Connections) Create() revel.Result {
  return c.Render()
}
