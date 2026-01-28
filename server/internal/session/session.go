package session

// 会话数据结构。

import "time"

type Data struct {
  UserID    string    `json:"user_id"`
  CSRFToken string    `json:"csrf_token"`
  CreatedAt time.Time `json:"created_at"`
  LastSeen  time.Time `json:"last_seen"`
}
