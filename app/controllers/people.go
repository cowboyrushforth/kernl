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
  target_user, err := models.UserFromSlug(rc,slug)
  if err != nil {
    return c.NotFound("person not found")
  }

  is_self := false
  connected_inbound := false
  connected_outbound := false
  if c.current_user().AccountIdentifier == target_user.AccountIdentifier {
    is_self = true
  } else {
    if target_user.SharesWithUser(rc, c.current_user().AccountIdentifier) {
      connected_inbound = true
    }
    if c.current_user().SharesWithUser(rc, target_user.AccountIdentifier) {
      connected_outbound = true
    }
  }
  return c.Render(target_user, is_self, connected_inbound, connected_outbound)
}

func (c People) ShowRemote(guid string) revel.Result {
  // get redis handle
  rc := GetRedisConn()
  defer rc.Close()
  person, err := models.PersonFromGuid(rc,guid)
  if err != nil {
    return c.NotFound("person not found")
  }
  connected_inbound := false
  connected_outbound := false
  if c.current_user().IsSharedWithByUser(rc, person.AccountIdentifier) {
    connected_inbound = true
  }
  if c.current_user().SharesWithUser(rc, person.AccountIdentifier) {
    connected_outbound = true
  }
  return c.Render(person, connected_inbound, connected_outbound)
}
