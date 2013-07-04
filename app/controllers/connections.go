package controllers

import "github.com/robfig/revel"
import "github.com/cowboyrushforth/go-webfinger"
import "github.com/cowboyrushforth/go-webfinger/jrd"
import "kernl/app/models"
import "net/http"
import "io/ioutil"

type Connections struct {
  Kernl
}

func (c Connections) checkUser() revel.Result {
  if user := c.current_user(); user == nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(Kernl.Index)
  }
  return nil
}

func (c Connections) Index() revel.Result {
  return c.Render()
}

func (c Connections) New() revel.Result {
  return c.Render()
}

func (c Connections) Verify(q string) revel.Result {
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
  // XXX: check validation
  // send salmon message to remote party

  success := false
  err := person.Connect(c.current_user()) 
  if err == nil {
    success = true
  }

  // if sharing message goes thru save connection
  return c.Render(success)
}
