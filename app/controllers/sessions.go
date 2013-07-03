package controllers

import "github.com/robfig/revel"
import "code.google.com/p/go.crypto/bcrypt"
import "kernl/app/models"
import "fmt"

type Sessions struct {
        *revel.Controller
}

func (c Sessions) New() revel.Result {
  return c.Render()
}

func (c Sessions) Create(Email string, Password string) revel.Result {
        // TODO Validations
        uid := fmt.Sprintf("user:%s", Email) 
        u, err := models.FetchUid(uid)
        if err == nil {
                errb := bcrypt.CompareHashAndPassword(u.PwdHash, []byte(Password))
                if errb == nil {
                        c.Session["uid"] = uid
                        c.Flash.Success("Welcome, " + u.Email)
                        return c.Redirect(Home.Index)
                } 
        }

        c.Flash.Out["username"] = Email
        c.Flash.Error("Login Failed")
        return c.Redirect(Sessions.New)
}

func (c Sessions) Destroy() revel.Result {
  c.Session["uid"] = ""
  return c.Redirect(Kernl.Index)
}
