package models

import "kernl/app/lib/redishandle"
import "github.com/garyburd/redigo/redis"
import "github.com/robfig/revel"
import "time"
import "errors"

type Notification struct {
  Id string
  Owner string
  Title string
  Message string
  Link string
  Read bool
}

func NotificationFromId(id string) (*Notification, error) {
  notification := Notification{}
  v, errb := redis.Values(redishandle.Do("HGETALL", id))
  if errb != nil {
    return nil, errb
  }
  errc := redis.ScanStruct(v, &notification)
  if errc != nil {
    return nil, errc
  }
  if len(notification.Id) == 0 {
    return nil, errors.New("notification not found")
  }

  return &notification, nil
}

func SendNotification(user *User, flavor string, guid string) {
  host_prefix := revel.Config.StringDefault("host.prefix", "http://localhost:9000")
  result, err := redishandle.Do("GET", redis.Args{}.Add("guid:"+guid)...)
  if err != nil {
    panic(err)
  }
  identifier, _ := redis.String(result, nil)
  if identifier == "" {
    panic("identifier not found")
  }

  mode := ""
  var person *Person = nil
  var local_user *User = nil
  if identifier[:5] == "user:" {
    mode = "user"
    local_user, _ = UserFromUid(identifier)
  } else if identifier[:7] == "person:" {
    mode = "person"
    person, _ = PersonFromUid(identifier)
  }

  notification := Notification{}
  notification.Read = false
  notification.Id = "notification:"+RandomString(16)
  notification.Owner = user.AccountIdentifier
  notification.Title = "Connections"
  switch flavor {
  case "share_started":
    if mode == "user" {
      notification.Message = local_user.DisplayName+" has started sharing with you"
      notification.Link = host_prefix+"/u/"+local_user.Slug
    } else if mode == "person" {
     notification.Message = person.DisplayName+" has started sharing with you"
     notification.Link = host_prefix+"/r/"+person.RemoteGuid
    }
  default:
    revel.INFO.Println("flavor", flavor, "not understood")
    return
  }
  notification.Insert(user)
}
func (self *Notification) Upsert() {
  _, errb := redishandle.Do("HMSET", redis.Args{}.Add(self.Id).AddFlat(self)...)
  if errb != nil {
    panic(errb)
  }
}

func (self *Notification) MarkAsRead(user *User) {
  self.Read = true
  self.Upsert()
  _, errd := redishandle.Do("HINCRBY", redis.Args{}.Add("user:"+user.Slug).
                             Add("NotificationCount").
                             Add(-1)...)

  if errd != nil {
    panic(errd)
  }
}

func (self *Notification) Insert(user *User) bool {
  self.Upsert()

  _, errc := redishandle.Do("ZADD", redis.Args{}.Add(user.NotificationsKey()).
                          Add(int32(time.Now().Unix())).
                          Add(self.Id)...)

  if errc != nil {
    panic(errc)
  }

  _, errd := redishandle.Do("HINCRBY", redis.Args{}.Add("user:"+user.Slug).
                             Add("NotificationCount").
                             Add(1)...)

  if errd != nil {
    panic(errd)
  }

  return true
}

func ListCurrentNotifications(user *User) []*Notification {
    result, errb := redishandle.Do("ZRANGE", redis.Args{}.Add(user.NotificationsKey()).Add(0).Add(-1)...)
    if errb != nil {
      panic(errb)
    }
    identifiers, _ := redis.Strings(result, nil)
    return materializeNotificationList(identifiers, false)
}

func materializeNotificationList(identifiers []string, show_read bool) []*Notification { 
    notifications := []*Notification{}
    for _,element := range identifiers {
      notification, err := NotificationFromId(element)
      if err == nil {
        if show_read == true || notification.Read == false {
          notifications = append(notifications, notification)
        }
      }
    }
    return notifications
}
