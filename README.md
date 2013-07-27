Kernl
================

This is an experimental project to attempt to build
a distributed social networking engine, compatible with
pump.io, diaspora, and hopefully others.

It's goal is to remain simple, and do things in a less is more fashion.

It is largely an attempt to learn both Golang, and the [Revel](http://robfig.github.io/revel)
web framework.  It is very much in a rough draft proof-of-concept state right now.

It is completely unfinished, untested, and largely a slow moving
work-in-progress, hopefully someday would be far enough along to participate
in federation in many different kinds of message exchanges.

This started as a diaspora centric project, but the main focus has since
shifted to being pump.io centric, however it would be cool to support
them all, so the diaspora code has been left in.  There will be feasibility
challenges that may arise that this project aims to explore and solve.

Patches, Feedback, Forks, Ideas, Etc all welcome.


Todo
-------------------
* basic test suite setup
* finish basic following support
* basic post and receive post support
** comments
** likes
* avatar/gravatar/etc support

Fire up
------------------
1. follow this tutorial http://robfig.github.io/revel/tutorial/createapp.html
2. once you get the tutorial working you can clone this repository into GOPATH/src/ so its at GOPATH/src/kernl
3. copy conf/app.conf.dist to conf/app.conf
3. revel run kernl should fire it up on port 9000

Gotchas
------------------
* if you move hostnames around things will break. ie - if your foobar@xyz.com:9000, and you change it to port 80, your handle ultimately changes.
* to use ssl, it is probably easiest to put it behind nginx or lighttpd.




