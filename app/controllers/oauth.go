package controllers

import "github.com/robfig/revel"
import "kernl/app/models"
import "github.com/cowboyrushforth/goa1"

type Oauth struct {
  *revel.Controller
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
