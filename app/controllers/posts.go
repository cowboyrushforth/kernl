package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "time"

type Posts struct {
  KernlAuthed
}

func (c Posts) Create(message string) revel.Result {
  // TODO: validation
  post := models.Post{
    Message: message,
    Guid: models.RandomString(32),
    AccountIdentifier: c.current_user().AccountIdentifier,
    Public: true,
    CreatedAt: time.Now().Unix(),
    DisplayName: c.current_user().DisplayName,
  }
  // sender is ourself
  sender := c.current_user().Person()
  post.Insert(sender)
  c.Flash.Success("Post Sent")
  return c.Redirect(Home.Index)
}
