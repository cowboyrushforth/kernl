package controllers

import "github.com/robfig/revel"

type Content struct {
        *revel.Controller
}

func (c Content) About() revel.Result {
  return c.Render()
}
