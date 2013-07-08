package models

import "github.com/garyburd/redigo/redis"
import "time"
import "errors"
import "github.com/robfig/revel"
import "github.com/dustin/go-humanize"


type Photo struct {
  Guid string
  AccountIdentifier string
  Public bool
  CreatedAt int64
  RemotePhotoPath string
  RemotePhotoName string
  PostGuid string
  Height int
  Width int
}

type Post struct {
  DisplayName string
  Message string
  Guid string
  AccountIdentifier string
  Public bool
  CreatedAt int64
  Likes int
  Dislikes int
  Photo Photo `redis:"-"`
}

func (self *Post) Id() string {
  return "post:"+self.Guid
}

func PostFromId(c redis.Conn, id string) (*Post, error) {
  post := Post{}
  v, errb := redis.Values(c.Do("HGETALL", id))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &post)
  if errc != nil {
    return nil, errc
  }
  if len(post.AccountIdentifier) == 0 {
    return nil, errors.New("post not found")
  }
  return &post, nil
}

func (self *Post) Insert(c redis.Conn, sender *Person) bool {
  // sanity check so we only insert or upsert ourselves
  result, err := c.Do("GET", redis.Args{}.Add("guid:"+self.Guid)...)
  if err != nil {
    panic(err)
  }
  identifier, _ := redis.String(result, nil)
  if identifier == "" {
    _, erra := c.Do("SET", redis.Args{}.Add("guid:"+self.Guid).Add(self.Id())...)
    if erra != nil {
      panic("data access problem")
    }
    identifier = self.Id()
  }

  // if its a new user,
  // or if it matches a person
  if identifier == self.Id() {
    _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
    if errb != nil {
      panic(errb)
    }


    // ok add this to the senders list of posts
    _, errc := c.Do("ZADD", redis.Args{}.Add(sender.PostsKey()).
    Add(time.Now().Unix()).Add(self.Id())...)
    if errc != nil {
      panic(errc)
    }

    // ok add this to their receivers
    self.DistributeToReceivers(c, sender)

    return true
  } 

  return false
}

func (self *Post) DistributeToReceivers(c redis.Conn, sender *Person) {
  user, user_err := UserFromGuid(c, sender.RemoteGuid)
  results, errb := c.Do("ZRANGE", redis.Args{}.Add(sender.ConnectionsOutboundKey()).Add(0).Add(-1)...)
  if errb != nil {
    panic(errb)
  }
  identifiers, _ := redis.Strings(results, nil)
  identifiers = append(identifiers, sender.AccountIdentifier)
  for _,identifier := range identifiers {
    revel.INFO.Println("distributing to", identifier)
    if IsLocalIdentifier(identifier) {
      revel.INFO.Println("\tLOCAL")
      _, errd := c.Do("ZADD", redis.Args{}.Add("feed:"+identifier).
      Add(time.Now().Unix()).Add(self.Id())...)
      if errd != nil {
        panic(errd)
      }
    } else if user_err == nil {
      revel.INFO.Println("\tNOT LOCAL", identifier)
      recipient,err := PersonFromUid(c, "person:"+identifier)
      if err == nil {
        revel.INFO.Println("\t\tHRM")
        result, r_err := SendStatusMessage(user, recipient, self) 
        if r_err != nil {
          panic(r_err)
        }
        if result.StatusCode != 200 && result.StatusCode != 202 {
          panic(result.StatusCode)
        }
      }
    } else {
      revel.INFO.Println("\tUSER NOT FOUND", sender.RemoteGuid, "x")
    }
  }
}

func (self *Post) HumanTime() string {
  return humanize.Time(time.Unix(self.CreatedAt, 0))
}
