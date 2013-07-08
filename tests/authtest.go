package tests

import "github.com/robfig/revel"

type AuthTest struct {
	revel.TestSuite
}

func (t AuthTest) Before() {
	println("Set up")
}

func (t AuthTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html")
}

func (t AuthTest) After() {
	println("Tear down")
}
