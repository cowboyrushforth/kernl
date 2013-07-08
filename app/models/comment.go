package models

import "github.com/garyburd/redigo/redis"
import "time"
import "errors"
//import "github.com/robfig/revel"
import "github.com/dustin/go-humanize"
import "kernl/app/lib/redishandle"

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

func CommentFromId(id string) (*Comment, error) {
  comment := Comment{}
  v, errb := redis.Values(redishandle.Do("HGETALL", id))
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

func (self *Comment) ParentPost() (*Post, error) {
  return PostFromId(self.ParentId())
}

func (self *Comment) Insert(sender *Person) bool {
  // sanity check so we only insert or upsert ourselves
  result, err := redishandle.Do("GET", redis.Args{}.Add("guid:"+self.Guid)...)
  if err != nil {
    panic(err)
  }
  identifier, _ := redis.String(result, nil)
  if identifier == "" {
    _, erra := redishandle.Do("SET", redis.Args{}.Add("guid:"+self.Guid).Add(self.Id())...)
    if erra != nil {
      panic("data access problem")
    }
    identifier = self.Id()
  }

  // if its a new comment,
  // or if it matches a person
  if identifier == self.Id() {
    post, errp := self.ParentPost()
    if errp == nil {
      _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id()).AddFlat(self)...)
      if errb != nil {
        panic(errb)
      }

      // ok add the parent to the senders list of posts
      _, errc := redishandle.Do("ZADD", redis.Args{}.Add(sender.PostsKey()).
      Add(time.Now().Unix()).Add(self.ParentId())...)
      if errc != nil {
        panic(errc)
      }

      // add comment to post comments set
      _, errd := redishandle.Do("ZADD", redis.Args{}.Add(post.CommentsKey()).
      Add(time.Now().Unix()).Add(self.Id())...)
      if errd != nil {
        panic(errd)
      }

      // ok add this to their receivers
      post.DistributeToReceivers(sender)

      return true
    }
  } 

  return false
}

func (self *Comment) HumanTime() string {
  return humanize.Time(time.Unix(self.CreatedAt, 0))
}
