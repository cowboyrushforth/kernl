package controllers

import "code.google.com/p/go.crypto/bcrypt"
import "github.com/robfig/revel"
import "kernl/app/models"
import "strings"

type Users struct {
  Kernl
}

func (c Users) New() revel.Result {
  return c.Render()
}

func (c Users) Create(Slug string,
                      Email string, 
                      Password string, 
                      PasswordConfirmation string) revel.Result {

      // get redis handle
      rc := GetRedisConn()
      defer rc.Close()

      user := models.User{}
      user.Slug = strings.ToLower(Slug)
      user.Email = strings.ToLower(Email)

      // add validation for password
      c.Validation.Required(PasswordConfirmation)
      c.Validation.Required(PasswordConfirmation == Password).Message("Password does not match")

      // validate user model
      user.Validate(rc, c.Validation)

      // shows errs if any
      if c.Validation.HasErrors() {
          c.Validation.Keep()
          c.FlashParams()
          return c.Redirect(Users.New)
      }

      user.PwdHash, _ = bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
      err := user.Insert(rc)
      if err != true {
        panic("oh no")
      }

      c.Session["user"] = user.DisplayName
      c.Session["uid"]  = user.Id()
      c.Flash.Success("Welcome, " + user.Email)
      return c.Redirect(Kernl.Index)
}
