package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Home struct {
  KernlAuthed
}

func (c Home) Index() revel.Result {
  notifications := models.ListCurrentNotifications(c.current_user())
  feed := models.HomeFeedForUser(c.current_user())
  return c.Render(notifications, feed)
}

