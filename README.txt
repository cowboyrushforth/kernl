[7/3/13 9:46:43 AM] Scott Rushforth: then put this in $GOPATH/src/kernl
[7/3/13 9:46:52 AM] Scott Rushforth: then revel run kernl
[7/3/13 9:46:57 AM] Scott Rushforth: and http localhost 9000 should come alive
[7/3/13 9:47:01 AM] Scott Rushforth: redis has to be running too

Install Go: https://code.google.com/p/go/downloads/list

Install Mercurial: http://mercurial.selenic.com/

export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
export GOPATH=/Users/me/code

cd /Users/me/code

go get github.com/robfig/revel/revel

ls should show:

bin pkg src

./bin/revel run github.com/robfig/revel/samples/chat
http://localhost:9000/ should work
Control-C

cd src
git clone git@tan.webair.com:kernl.git
cd ..
go get code.google.com/p/go.crypto/bcrypt
go get github.com/cowboyrushforth/go-webfinger

./bin/revel run kernl
http://localhost:9000/ should work


Coffee:

install node: http://nodejs.org/download/
sudo npm install -g coffee-script

-------------------- 
Ref Sites:

https://github.com/robfig/revel
http://robfig.github.io/revel/
http://mercurial.selenic.com/
https://code.google.com/p/go/downloads/list
http://golang.org/

http://nodejs.org/download/
http://coffeescript.org/

