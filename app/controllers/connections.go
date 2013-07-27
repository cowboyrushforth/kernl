package controllers

import "github.com/robfig/revel"
import "github.com/cowboyrushforth/go-webfinger"
import "github.com/cowboyrushforth/go-webfinger/jrd"
import "kernl/app/models"
import "net/http"
import "io/ioutil"

type Connections struct {
  KernlAuthed
}

func (c Connections) Index() revel.Result {
  user := c.current_user()
  list := user.ListConnections( true, true)
  return c.Render(list)
}

func (c Connections) Following() revel.Result {
  user := c.current_user()
  list := user.ListConnections(true, false)
  return c.Render(list)
}

func (c Connections) Followers() revel.Result {
  user := c.current_user()
  list := user.ListConnections(false, true)
  return c.Render(list)
}

func (c Connections) Blocked() revel.Result {
  user := c.current_user()
  list := user.ListBlockedConnections()
  return c.Render(list)
}

func (c Connections) Mutual() revel.Result {
  user := c.current_user()
  list := user.ListMutualConnections()
  return c.Render(list)
}

func (c Connections) New() revel.Result {
  return c.Render()
}

func (c Connections) Verify(q string) revel.Result {
  // see if we have this connection already
  if c.current_user().HasOutboundConnection(q) {
    c.Flash.Error("You are already connected to "+q)
    return c.Redirect(Connections.Index)
  }

  revel.INFO.Println("about to finger:", q)

  client := webfinger.NewClient(nil)

  insecure_fingering := revel.Config.BoolDefault("auth.allow_unsecure_fingering", false)

  err := error(nil)
  resource := &jrd.JRD{}
  if insecure_fingering {
    resource, err = client.LookupInsecure(q, []string{})
  } else {
    resource, err = client.Lookup(q, []string{})
  }

  if err != nil {
    has_errors := true
    return c.Render(has_errors)
  }

  subject := resource.Subject

  profile_href := ""
  profile_link := resource.GetLinkByRel("http://webfinger.net/rel/profile-page")
  if profile_link != nil {
    profile_href = profile_link.Href
  }

  hcard_href := ""
  hcard_link := resource.GetLinkByRel("http://microformats.org/profile/hcard")
  if hcard_link != nil {
    hcard_href = hcard_link.Href
  }

  seed_href := ""
  seed_link := resource.GetLinkByRel("http://joindiaspora.com/seed_location")
  if seed_link != nil {
    seed_href = seed_link.Href
  }

  guid := ""
  guid_link := resource.GetLinkByRel("http://joindiaspora.com/guid")
  if guid_link != nil {
    guid = guid_link.Href
  }

  hcard_bytes := []byte{}
  if hcard_href != "" {
    resp, err := http.Get(hcard_href)
    if err != nil {
        // handle error
      }
      defer resp.Body.Close()
      hcard_bytes, err = ioutil.ReadAll(resp.Body)
  }
  hcard_body := string(hcard_bytes)

  pubkey := ""
  pubkey_link := resource.GetLinkByRel("diaspora-public-key")
  if pubkey_link != nil {
    pubkey = pubkey_link.Href
  }

  // render all of the info into a template
  return c.Render(subject, profile_href, hcard_href, hcard_body, seed_href, guid, pubkey)
}

func (c Connections) Create(person models.Person) revel.Result {

  // validate user model
  person.Validate(c.Validation)

  // shows errs if any
  if c.Validation.HasErrors() {
    c.Validation.Keep()
    c.FlashParams()
    return c.RenderText("did not pass validation")
  }

  // send salmon message to remote party
  // and save Person if successful
  success := false
  err := person.Connect(c.current_user()) 
  if err == nil {
    success = true
  }

  // if sharing message goes thru save connection
  return c.Render(success, person)
}
