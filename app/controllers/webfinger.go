package controllers

import "github.com/robfig/revel"

type Webfinger struct {
  Kernl
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
  // TODO: use server config for hostname
  link := Link{Rel: "ldd", 
               Template: "http://localhost:9000/wf?q={uri}",
               Xtype: "application/xrd+xml",
               Title: "Webfinger Template"}

  xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
             Xmlnshm: "http://host-meta.net/xrd/1.0",
             Link: link}

  return c.RenderXml(xrd)
}

func (c Webfinger) Show() revel.Result {
  // TODO: use server config for hostname
  // TODO: lookup email, print out slug

  link := Link{Rel: "http://microformats.org/profile/hcard",
              Href: "http://localhost:9000/users/scott"}

  xrd := XRD{Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
             Alias: "http://localhost:9000/users/scott",
             Link: link}
  return c.RenderXml(xrd)
}
