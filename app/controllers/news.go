package controllers

import "github.com/robfig/revel"

type News struct {
  Kernl
}

func (c News) Like(q string) revel.Result {
  return nil
}

