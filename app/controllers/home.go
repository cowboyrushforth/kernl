package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Home struct {
  KernlAuthed
}

func (c Home) Index() revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  notifications := models.ListCurrentNotifications(rc, c.current_user())
  feed := models.HomeFeedForUser(rc, c.current_user())
  return c.Render(notifications, feed)
}

