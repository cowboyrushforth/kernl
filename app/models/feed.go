package models

import "github.com/garyburd/redigo/redis"

func HomeFeedForUser(c redis.Conn, user *User) []*Post {
  results, err := c.Do("ZRANGE", redis.Args{}.Add("feed:"+user.AccountIdentifier)...)
  if err != nil {
    panic(err)
  }
  posts := []*Post{}
  identifiers, _ := redis.Strings(results, nil)
  for _,identifier := range identifiers {
    post, errb := PostFromId(c,identifier)
    if errb != nil {
      posts = append(posts, post)
    }
  }
  return posts
}
