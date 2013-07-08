package controllers

import "github.com/robfig/revel"
//import "kernl/app/models"

type Posts struct {
  KernlAuthed
}

func (c Posts) Create(message string) revel.Result {

      // get redis handle
      //rc := GetRedisConn()
      //defer rc.Close()

      //user := models.Post{}

//    c.Flash.Error("Please log in first")
     c.Flash.Success("Post Sent")
      return c.Redirect(Home.Index)
}
