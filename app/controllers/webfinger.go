package controllers

import "github.com/robfig/revel"

type Webfinger struct {
  *revel.Controller
}

type Link struct {
  Rel       string `xml:"rel,attr"`
  Template  string `xml:"template,attr,omitempty"`
  Xtype     string `xml:"type,attr,omitempty"`
  Href      string `xml:",omitempty"`
  Title     string `xml:",omitempty"`
}

type XRD struct {
  Xmlns   string `xml:"xmlns,attr"`
  Xmlnshm string `xml:"xmlns:hm,attr,omitempty"`
  Alias   string `xml:",omitempty"`
  Subject   string `xml:",omitempty"`
  Link    Link
}

func (c Webfinger) Index() revel.Result {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")

  link := Link{Rel: "lrdd", 
               Template: host_prefix+"/wf?q={uri}",
               Xtype: "application/xrd+xml",
               Title: "Webfinger Template"}

  xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
             Xmlnshm: "http://host-meta.net/xrd/1.0",
             Link: link}

  return c.RenderXml(xrd)
}

func (c Webfinger) Show() revel.Result {
  // TODO: lookup email, print out slug
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")

  link := Link{Rel: "http://microformats.org/profile/hcard",
              Href: host_prefix+"/users/scott"}

  xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
             Alias: host_prefix+"/users/scott",
             Link: link}
  return c.RenderXml(xrd)
}
