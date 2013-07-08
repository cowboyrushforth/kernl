package controllers

import "github.com/robfig/revel"
import "github.com/thomas11/atomgenerator"
import "kernl/app/models"

type Public struct {
  Kernl
}

func (c Public) Index(slug string) revel.Result {
  user, err := models.UserFromSlug(slug)
  if err != nil {
    return c.NotFound("user not found")
  }
  f :=  atomgenerator.Feed{
    Title:   user.DisplayName+"'s Feed",
    Link:    "http://www.myblog.bogus",
  }
  f.AddAuthor(atomgenerator.Author{
        Name: "",
        Uri: "",
  })
  atom, _  := f.GenXml()
  return c.RenderText(string(atom))
}

func (c Public) PublicFeed() revel.Result {
  return c.RenderJson([]string{})
}
