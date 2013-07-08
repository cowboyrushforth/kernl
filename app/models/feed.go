package models

import "github.com/garyburd/redigo/redis"
import "kernl/app/lib/redishandle"

func HomeFeedForUser(user *User) []*Post {
  start := 0
  limit := 20
  results, err := redishandle.Do("SORT", redis.Args{}.Add("feed:"+user.AccountIdentifier).
                        Add("LIMIT").Add(start).Add(limit).
                        Add("BY").Add("*->CreatedAt").
                        Add("DESC")...)
  if err != nil {
    panic(err)
  }
  posts := []*Post{}
  identifiers, _ := redis.Strings(results, nil)
  for _,identifier := range identifiers {
    post, errb := PostFromId(identifier)
    if errb != nil {
      panic(errb)
    }
    posts = append(posts, post)
  }
  return posts
}
