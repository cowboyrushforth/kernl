package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "github.com/cowboyrushforth/goa1"
import "net/url"
import "strings"

type Oauth struct {
  Kernl
}

func (c Oauth) RequestToken() revel.Result {
  req, erra := goa1.ParseRequest(c.Request.Request) 
  if erra != nil {
    panic(erra)
  }

  client,errb := models.ClientFromConsumerKey(req.ConsumerKey)
  if errb != nil {
    panic(errb)
  }

  ok, errc := goa1.Validate(req, client.Secret, "")
  if errb != nil {
    panic(errc)
  }

  if ok {
    revel.INFO.Println("OK")
  } else {
    revel.INFO.Println("NOT OK")
    c.Response.Status = 400
    return c.RenderText("FAIL")
  }

  request_token := models.RequestToken{
    ConsumerKey: req.ConsumerKey,
    Callback: req.Callback}

  if request_token.Insert() == false {
    panic("data storage error")
  }

  reply := goa1.OAuthRequestTokenReply{
    CallbackConfirmed: true,
    Token: request_token.Token,
    TokenSecret: request_token.TokenSecret}

  return c.RenderText(goa1.RequestTokenPayload(&reply))
}

func (c Oauth) Authorize(oauth_token string, verifier string) revel.Result {
  if c.current_user() == nil {
    return c.Redirect("/login?return_to="+url.QueryEscape(c.Request.Request.URL.String()))
  }
  request_token, err := models.RequestTokenFromToken(oauth_token)
  if err != nil {
    panic(err)
  }
  if request_token.Used  {
    c.Response.Status = 400
    return c.RenderText("Token Invalid")
  }
  if len(request_token.AccountIdentifier) > 0 && 
    (request_token.AccountIdentifier != c.current_user().AccountIdentifier) {
    c.Response.Status = 400
    return c.RenderText("Token Invalid")
  }
  client, errb := models.ClientFromConsumerKey(request_token.ConsumerKey)
  if errb != nil {
    panic(errb)
  }

  request_token.AccountIdentifier = c.current_user().AccountIdentifier
  if request_token.Insert() != true {
    panic("data storage error")
  }

  // OK Access Granted
  if verifier != "" {
    if verifier != request_token.Verifier {
      panic("verifier problem")
    }
     request_token.Used = true
     request_token.Authenticated = true
     if request_token.Insert() == false {
       panic("data storage error")
     }

     access_token := models.AccessToken{
       AccountIdentifier: request_token.AccountIdentifier,
       RequestToken: request_token.Token,
       ConsumerKey: request_token.ConsumerKey}

     if access_token.Insert() == false {
       panic("data storage error")
     }

     if request_token.Callback != "" {
       request_token.Callback, _ = url.QueryUnescape(request_token.Callback)
       if strings.Contains(request_token.Callback, "?") {
         return c.Redirect(request_token.Callback+"&oauth_token="+request_token.Token+"&oauth_verifier="+request_token.Verifier)
       } else {
         return c.Redirect(request_token.Callback+"?oauth_token="+request_token.Token+"&oauth_verifier="+request_token.Verifier)
       }
     } else {
       return c.RenderText("NOT IMPLEMENTED FULLY YET")
     }
  }

  return c.Render(request_token, client)
}
