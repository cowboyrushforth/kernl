package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type People struct {
  KernlAuthed
}

func (c People) Show(slug string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  target, err := models.UserFromSlug(rc,slug)
  if err != nil {
    return c.NotFound("person not found")
  }

  is_self := false
  connected_inbound := false
  connected_outbound := false
  if c.current_user().AccountIdentifier == target.AccountIdentifier {
    is_self = true
  } else {
    if target.SharesWithUser(rc, c.current_user().AccountIdentifier) {
      connected_inbound = true
    }
    if c.current_user().SharesWithUser(rc, target.AccountIdentifier) {
      connected_outbound = true
    }
  }
  return c.Render(target, is_self, connected_inbound, connected_outbound)
}

func (c People) ShowRemote(guid string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  target, err := models.PersonFromGuid(rc,guid)
  if err != nil {
    return c.NotFound("person not found")
  }
  connected_inbound := false
  connected_outbound := false
  if c.current_user().IsSharedWithByUser(rc, target.AccountIdentifier) {
    connected_inbound = true
  }
  if c.current_user().SharesWithUser(rc, target.AccountIdentifier) {
    connected_outbound = true
  }
  return c.Render(target, connected_inbound, connected_outbound)
}
