package models

import "github.com/garyburd/redigo/redis"
import "time"
import "errors"
//import "github.com/robfig/revel"
import "github.com/dustin/go-humanize"

type Comment struct {
  DisplayName string
  Text string
  Guid string
  ParentGuid string
  AccountIdentifier string
  CreatedAt int64
  Likes int
  Dislikes int
}

func CommentFromId(c redis.Conn, id string) (*Comment, error) {
  comment := Comment{}
  v, errb := redis.Values(c.Do("HGETALL", id))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &comment)
  if errc != nil {
    return nil, errc
  }
  if len(comment.AccountIdentifier) == 0 {
    return nil, errors.New("post not found")
  }
  return &comment, nil
}

func (self *Comment) Id() string {
  return "comment:"+self.Guid
}

func (self *Comment) ParentId() string {
  return "post:"+self.ParentGuid
}

func (self *Comment) ParentPost(rc redis.Conn) (*Post, error) {
  return PostFromId(rc, self.ParentId())
}

func (self *Comment) Insert(c redis.Conn, sender *Person) bool {
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

  // if its a new comment,
  // or if it matches a person
  if identifier == self.Id() {
    post, errp := self.ParentPost(c)
    if errp == nil {
      _, errb := c.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
      if errb != nil {
        panic(errb)
      }

      // ok add the parent to the senders list of posts
      _, errc := c.Do("ZADD", redis.Args{}.Add(sender.PostsKey()).
      Add(time.Now().Unix()).Add(self.ParentId())...)
      if errc != nil {
        panic(errc)
      }

      // add comment to post comments set
      _, errd := c.Do("ZADD", redis.Args{}.Add(post.CommentsKey()).
      Add(time.Now().Unix()).Add(self.Id())...)
      if errd != nil {
        panic(errd)
      }

      // ok add this to their receivers
      post.DistributeToReceivers(c, sender)

      return true
    }
  } 

  return false
}

func (self *Comment) HumanTime() string {
  return humanize.Time(time.Unix(self.CreatedAt, 0))
}
