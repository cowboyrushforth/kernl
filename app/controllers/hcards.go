package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Hcards struct {
  *revel.Controller
}

func (c Hcards) Show(slug string) revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  host_suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")
  card, err := models.UserFromSlug(slug)
  if err != nil {
    return c.NotFound("Card Not Found")
  }

  return c.Render(host_prefix, host_suffix, card)
}
