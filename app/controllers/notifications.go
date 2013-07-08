package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Notifications struct {
  KernlAuthed
}

func (c Notifications) Ack(id string) revel.Result {
  // get redis handle
  notification, err := models.NotificationFromId(id)
  if err != nil || notification.Owner != c.current_user().AccountIdentifier {
    c.Response.Status = 400
    return c.RenderText("")
  }

  notification.MarkAsRead(c.current_user())

  return c.Redirect(Home.Index)
}

