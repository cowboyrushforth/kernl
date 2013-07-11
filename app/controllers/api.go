package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "github.com/cowboyrushforth/goa1"
import "net/http"
import "encoding/json"
import "io"
import "bytes"

type Api struct {
  Kernl
}

type FeedPost struct {
   Verb string `json:"verb"`
   ActivityObject models.ActivityObject `json:"object"`
}

func (c Api) IsVerifiedRequest(http *http.Request) bool {
  req, erra := goa1.ParseRequest(http) 
  if erra != nil {
    panic(erra)
  }

  client,errb := models.ClientFromConsumerKey(req.ConsumerKey)
  if errb != nil {
    panic(errb)
  }

  access_token, errc := models.AccessTokenFromToken(req.Token)
  ok, errc := goa1.Validate(req, client.Secret, access_token.TokenSecret)
  if errb != nil {
    panic(errc)
  }

  if ok  {
    c.RenderArgs["access_token"] = access_token
    return true
  }

  return false
}

func (c Api) Whoami() revel.Result {
  ok := c.IsVerifiedRequest(c.Request.Request)

  if ok {
    host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
    revel.INFO.Println("OK")
    return c.Redirect(host_prefix+"/api/user/"+c.RenderArgs["access_token"].(*models.AccessToken).Slug+"/profile")
  }

  revel.INFO.Println("NOT OK")
  c.Response.Status = 400
  return c.RenderText("FAIL")
}

func (c Api) Profile(slug string) revel.Result {
  if c.current_user() == nil {
    ok := c.IsVerifiedRequest(c.Request.Request)
    if ok == false {
      c.Response.Status = 400
      return c.RenderText("")
    }
  }
  user, err := models.UserFromSlug(slug)
  if err != nil {
    panic(err)
  }
  return c.RenderJson(user.ActivityObject())
}

func (c Api) FeedPost() revel.Result {
  var b bytes.Buffer
  var dest io.Writer = &b
  _,_ = io.Copy(dest, c.Request.Body)
  var payload FeedPost
  err := json.Unmarshal(b.Bytes(), &payload)
  if err != nil {
    panic(err)
  }
  revel.INFO.Println(payload)

  c.Response.Status = 400
  return c.RenderText("")
}
