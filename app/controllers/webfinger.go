package controllers

import "github.com/robfig/revel"
import "strings"
import "kernl/app/models"
import "encoding/base64"

type Webfinger struct {
  *revel.Controller
}

type Link struct {
  Rel       string `xml:"rel,attr"json:"rel,omitempty"`
  Template  string `xml:"template,attr,omitempty"json:"template,omitempty"`
  Xtype     string `xml:"type,attr,omitempty"json:"type,omitempty"`
  Href      string `xml:"href,attr,omitempty"json:"href,omitempty"`
  Title     string `xml:",omitempty"json:",omitempty"`
}

type XRD struct {
  Xmlns   string `xml:"xmlns,attr"`
  Xmlnshm string `xml:"xmlns:hm,attr,omitempty"`
  Alias   string `xml:",omitempty"`
  Subject string `xml:",omitempty"`
  Link    []Link
}

type JRD struct {
  Links []Link  `json:"links"`
}

func (c Webfinger) Index() revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")

  link := Link{Rel: "lrdd", 
               Template: host_prefix+"/wf?q={uri}",
               Xtype: "application/xrd+xml",
               Title: "Webfinger Template"}

  xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
             Xmlnshm: "http://host-meta.net/xrd/1.0",
             Link: []Link{link}}

  return c.RenderXml(xrd)
}

func (c Webfinger) IndexJson() revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")

  xml_lrdd := Link{Rel: "lrdd", 
  Template: host_prefix+"/wf?q={uri}",
  Xtype: "application/xrd+xml"}

  json_lrdd := Link{Rel: "lrdd", 
  Template: host_prefix+"/wf.json?q={uri}",
  Xtype: "application/json"}

  reg := Link{Rel: "registration_endpoint", 
  Href: host_prefix+"/api/client/register"}

  request := Link{Rel: "http://apinamespace.org/oauth/request_token", 
  Href: host_prefix+"/oauth/request_token"}

  authorize := Link{Rel: "http://apinamespace.org/oauth/authorize", 
  Href: host_prefix+"/oauth/authorize"}

  access := Link{Rel: "http://apinamespace.org/oauth/access_token", 
  Href: host_prefix+"/oauth/access_token"}

  dialback := Link{Rel: "dialback", 
  Href: host_prefix+"/api/dialback"}

  whoami := Link{Rel: "http://apinamespace.org/activitypub/whoami", 
  Href: host_prefix+"/api/whoami"}

  xrd := JRD{Links: []Link{xml_lrdd, json_lrdd, reg, request, authorize, access, dialback, whoami}}

  return c.RenderJson(xrd)
}


func (c Webfinger) Show(q string) revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  host_suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")

  if strings.Contains(q, "@"+host_suffix) {
    q = strings.Replace(q, "acct:", "", 1)
    slug := strings.Replace(q, "@"+host_suffix, "", 1)
    if len(slug) > 0 {
      user, err := models.UserFromSlug(slug)
      if err == nil {
        hcard := Link{Rel: "http://microformats.org/profile/hcard",
                     Xtype: "text/html",
                     Href: host_prefix+"/hcards/"+user.Slug}

        seed := Link{Rel: "http://joindiaspora.com/seed_location",
                     Xtype: "text/html",
                     Href: host_prefix}

        guid := Link{Rel: "http://joindiaspora.com/guid",
                     Xtype: "text/html",
                     Href: user.Guid}

        page := Link{Rel: "http://webfinger.net/rel/profile-page",
                     Xtype: "text/html",
                     Href: host_prefix+"/u/"+user.Slug}

        feed := Link{Rel: "http://schemas.google.com/g/2010#updates-from",
                     Xtype: "application/atom+xml",
                     Href: host_prefix+"/public/"+user.Slug+"/feed.atom"}

        key  := Link{Rel: "diaspora-public-key",
                     Xtype: "RSA",
                     Href: base64.StdEncoding.EncodeToString([]byte(user.RSAPubKey()))}

        xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
                   Alias: host_prefix,
                   Subject: "acct:"+q,
                   Link: []Link{hcard, seed, guid, page, feed, key}}

        return c.RenderXml(xrd)
      } else {
        revel.INFO.Println("could not find", slug)
      }
    }
  } 

  return c.NotFound("user not found")
}
