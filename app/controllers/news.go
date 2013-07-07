package controllers

import "github.com/robfig/revel"

type News struct {
  KernlAuthed
}

func (c News) Like(q string) revel.Result {
  x := []int{1, 2, 3, 4, 5}
  return c.RenderJson(x)
}

