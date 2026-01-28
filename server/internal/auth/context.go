package auth

// 用户上下文管理。

import "github.com/gin-gonic/gin"

const currentUserKey = "current_user"

type UserView struct {
  UserID string
  Email  string
}

func SetCurrentUser(c *gin.Context, user UserView) {
  c.Set(currentUserKey, user)
}

func CurrentUser(c *gin.Context) (*UserView, error) {
  val, ok := c.Get(currentUserKey)
  if !ok {
    return nil, ErrNotAuthenticated
  }
  user, ok := val.(UserView)
  if !ok {
    return nil, ErrNotAuthenticated
  }
  return &user, nil
}
