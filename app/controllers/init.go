package controllers

import "github.com/robfig/revel"

func init() {
        revel.OnAppStart(Init)
        revel.InterceptMethod(Home.checkUser, revel.BEFORE)
        revel.InterceptMethod(Connections.checkUser, revel.BEFORE)
}
