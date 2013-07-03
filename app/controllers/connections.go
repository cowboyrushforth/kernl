package controllers

import "github.com/robfig/revel"
import "github.com/cowboyrushforth/go-webfinger"
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

  resource, err := client.Lookup("r2d2@joindiaspora.com", []string{})
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



  return c.Render(subject, profile_href, hcard_href, hcard_body, seed_href, guid)
}

func (c Connections) Create() revel.Result {
  return c.Render()
}
