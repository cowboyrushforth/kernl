package controllers

import "github.com/robfig/revel"
import "strings"
import "kernl/app/models"
import "encoding/base64"

type Webfinger struct {
  *revel.Controller
}

type Link struct {
  Rel       string `xml:"rel,attr"`
  Template  string `xml:"template,attr,omitempty"`
  Xtype     string `xml:"type,attr,omitempty"`
  Href      string `xml:"href,attr,omitempty"`
  Title     string `xml:",omitempty"`
}

type XRD struct {
  Xmlns   string `xml:"xmlns,attr"`
  Xmlnshm string `xml:"xmlns:hm,attr,omitempty"`
  Alias   string `xml:",omitempty"`
  Subject string `xml:",omitempty"`
  Link    []Link
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

func (c Webfinger) Show(q string) revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  host_suffix := revel.Config.StringDefault("host.suffix", "localhost:9000")

  if strings.Contains(q, "@"+host_suffix) {
    slug := strings.Replace(q, "@"+host_suffix, "", 1)
    if len(slug) > 0 {
      // find user
      rc := GetRedisConn()
      defer rc.Close()
      user, err := models.UserFromSlug(rc, slug)
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
                     Href: host_prefix+"/public/"+user.Slug+".atom"}

        key  := Link{Rel: "diaspora-public-key",
                     Xtype: "RSA",
                     Href: base64.StdEncoding.EncodeToString([]byte(user.RSAPubKey()))}

        xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
                   Alias: host_prefix,
                   Subject: "acct:"+q,
                   Link: []Link{hcard, seed, guid, page, feed, key}}

        return c.RenderXml(xrd)
      }
    }
  } 

  return c.NotFound("user not found")
}
