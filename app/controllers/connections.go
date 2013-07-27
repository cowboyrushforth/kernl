package controllers

import "github.com/robfig/revel"
import "kernl/app/models"

type Connections struct {
  KernlAuthed
}

func (c Connections) Index() revel.Result {
  user := c.current_user()
  list := user.ListConnections( true, true)
  return c.Render(list)
}

func (c Connections) Following() revel.Result {
  user := c.current_user()
  list := user.ListConnections(true, false)
  return c.Render(list)
}

func (c Connections) Followers() revel.Result {
  user := c.current_user()
  list := user.ListConnections(false, true)
  return c.Render(list)
}

func (c Connections) Blocked() revel.Result {
  user := c.current_user()
  list := user.ListBlockedConnections()
  return c.Render(list)
}

func (c Connections) Mutual() revel.Result {
  user := c.current_user()
  list := user.ListMutualConnections()
  return c.Render(list)
}

func (c Connections) New() revel.Result {
  return c.Render()
}

func (c Connections) Verify(q string) revel.Result {
  // see if we have this connection already
  if c.current_user().Follows(q) {
    c.Flash.Error("You are already connected to "+q)
    return c.Redirect(Connections.Index)
  }

  person, err := models.PersonFromWebFinger(q)
  if err != nil {
    has_errors := true
    return c.Render(has_errors)
  }


  return c.Render(person)
}

func (c Connections) Create(ident string) revel.Result {
  person, err := models.PersonFromWebFinger(ident)
  if err != nil {
    panic(err)
  }

  // validate user model
  person.Validate(c.Validation)

  // shows errs if any
  if c.Validation.HasErrors() {
    c.Validation.Keep()
    c.FlashParams()
    return c.RenderText("did not pass validation")
  }

  success := false
  errb := person.Connect(c.current_user()) 
  if errb == nil {
    success = true
  }

  return c.Render(success, person)
}
