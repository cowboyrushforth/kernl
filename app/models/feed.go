package models

import "github.com/garyburd/redigo/redis"

func HomeFeedForUser(c redis.Conn, user *User) []*Post {
  start := 0
  limit := 20
  results, err := c.Do("SORT", redis.Args{}.Add("feed:"+user.AccountIdentifier).
                        Add("LIMIT").Add(start).Add(limit).
                        Add("BY").Add("*->CreatedAt").
                        Add("DESC")...)
  if err != nil {
    panic(err)
  }
  posts := []*Post{}
  identifiers, _ := redis.Strings(results, nil)
  for _,identifier := range identifiers {
    post, errb := PostFromId(c,identifier)
    if errb != nil {
      panic(errb)
    }
    posts = append(posts, post)
  }
  return posts
}
