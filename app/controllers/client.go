package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Client struct {
  *revel.Controller
}

type RegResult struct {
 ClientId string `json:"client_id"`
 ClientSecret string `json:"client_secret"`
 ExpiresAt int `json:"expires_at"`
}

func (c Client) Register() revel.Result {
  reg_type := c.Params.Get("type")
  if reg_type == "" ||
  reg_type != "client_update" &&
  reg_type != "client_associate" {
    c.Response.Status = 400
    return c.RenderText("Invalid registration type")
  }

  client := models.Client{}
  client_id     := c.Params.Get("client_id")
  access_token  := c.Params.Get("access_token")
  client_secret := c.Params.Get("client_secret")
  contacts      := c.Params.Get("contacts")
  apptype       := c.Params.Get("application_type")
  appname       := c.Params.Get("application_name")
  logo_url      := c.Params.Get("logo_url")
  request_uris  := c.Params.Get("request_uris")

  if access_token != "" {
    c.Response.Status = 400
    return c.RenderText("access_token not needed for registration")
  }
  if client_id != "" {
    if reg_type != "client_update" {
      c.Response.Status = 400
      return c.RenderText("Only set client_id for update.")
    }
    client.ConsumerKey = client_id
  }
  if client_secret != "" {
    if reg_type != "client_update" {
      c.Response.Status = 400
      return c.RenderText("Only set client_secret for update.")
    }
    client.Secret = client_secret
  }

  if contacts != "" {
    // XXX: fill in contacts
    // XXX: ensure each is a valid email
    // XXX: space separated
  }


  if apptype != "" {
    if apptype != "web" && apptype != "native" {
      c.Response.Status = 400
      return c.RenderText("Unknown application_type.")
    }
    client.XType = apptype
  }

  client.Title = appname

  if logo_url != "" {
    // XXX: if valid url fill in Client.LogoUrl
  }

  if request_uris != "" {
    // XXX: space separated list of urls
    // XXX: fill in Client.RequestURIs
  }

  /*
  if (req.remoteHost) {
    props.host = req.remoteHost;
  } else if (req.remoteUser) {
    props.webfinger = req.remoteUser;
  }
  */

  if reg_type == "client_associate" {
   if client.Insert() == false {
        panic("could not insert client")
   }
  } else if reg_type == "client_update" {
    panic("not implemented yet")
    // TODO lookup Client by consumer key
    // make sure client.Secret matches params Secret
    // update given props
    // render json client_id, client_secret, expires_at:0
  }

  reg_result := RegResult{
        ClientId: client.ConsumerKey,
        ClientSecret: client.Secret,
        ExpiresAt: 0}

  return c.RenderJson(reg_result)
}
