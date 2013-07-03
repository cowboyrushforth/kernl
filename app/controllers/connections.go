package controllers

import "github.com/robfig/revel"
import "github.com/ant0ine/go-webfinger"
import "fmt"

type Connections struct {
  Kernl
}

func (c Connections) checkUser() revel.Result {
  if user := c.current_user(); user == nil {
    c.Flash.Error("Please log in first")
    return c.Redirect(Kernl.Index)
  }
  return nil
}

func (c Connections) Index() revel.Result {
  return c.Render()
}

func (c Connections) New() revel.Result {
  return c.Render()
}

func (c Connections) Verify(q string) revel.Result {
  revel.INFO.Println("about to finger:", q)


  client := webfinger.NewClient(nil)

  resource, err := client.Lookup("scott@localhost:9000", []string{})
  if err != nil {
    panic(err)
  }
  fmt.Printf("JRD: %+v", resource)

  return c.Render()
}

func (c Connections) Create() revel.Result {
  return c.Render()
}
