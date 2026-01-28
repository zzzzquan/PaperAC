package store

// Redis连接与会话/频控存储。

import (
  "context"
  "encoding/json"
  "fmt"
  "time"

  "aigc-detector/server/internal/config"
  "aigc-detector/server/internal/session"

  "github.com/redis/go-redis/v9"
)

type RedisStore struct {
  Client *redis.Client
}

func NewRedis(cfg config.Config) (*RedisStore, error) {
  client := redis.NewClient(&redis.Options{
    Addr:     cfg.RedisAddr,
    Password: cfg.RedisPass,
    DB:       cfg.RedisDB,
  })
  if err := client.Ping(context.Background()).Err(); err != nil {
    return nil, err
  }
  return &RedisStore{Client: client}, nil
}

func (r *RedisStore) Close() {
  if r.Client != nil {
    _ = r.Client.Close()
  }
}

func (r *RedisStore) SetSession(ctx context.Context, sessionID string, data session.Data, ttl time.Duration) error {
  payload, err := json.Marshal(data)
  if err != nil {
    return err
  }
  return r.Client.Set(ctx, sessionKey(sessionID), payload, ttl).Err()
}

func (r *RedisStore) GetSession(ctx context.Context, sessionID string) (*session.Data, error) {
  val, err := r.Client.Get(ctx, sessionKey(sessionID)).Result()
  if err != nil {
    if err == redis.Nil {
      return nil, nil
    }
    return nil, err
  }
  var data session.Data
  if err := json.Unmarshal([]byte(val), &data); err != nil {
    return nil, err
  }
  return &data, nil
}

func (r *RedisStore) DeleteSession(ctx context.Context, sessionID string) error {
  return r.Client.Del(ctx, sessionKey(sessionID)).Err()
}

func (r *RedisStore) EmailCooldown(ctx context.Context, email string, ttl time.Duration) (bool, error) {
  key := fmt.Sprintf("rl:email:%s", email)
  ok, err := r.Client.SetNX(ctx, key, "1", ttl).Result()
  return ok, err
}

func (r *RedisStore) ClearEmailCooldown(ctx context.Context, email string) {
  key := fmt.Sprintf("rl:email:%s", email)
  _ = r.Client.Del(ctx, key).Err()
}

func (r *RedisStore) IncrementIPRate(ctx context.Context, ip string, windowKey string, ttl time.Duration) (int64, error) {
  key := fmt.Sprintf("rl:ip:%s:%s", ip, windowKey)
  val, err := r.Client.Incr(ctx, key).Result()
  if err != nil {
    return 0, err
  }
  if val == 1 {
    _ = r.Client.Expire(ctx, key, ttl).Err()
  }
  return val, nil
}

func sessionKey(sessionID string) string {
  return fmt.Sprintf("sess:%s", sessionID)
}

func (r *RedisStore) PushTask(ctx context.Context, queue string, taskID string) error {
  return r.Client.LPush(ctx, queue, taskID).Err()
}

func (r *RedisStore) PopTask(ctx context.Context, queue string, timeout time.Duration) (string, error) {
  result, err := r.Client.BRPop(ctx, timeout, queue).Result()
  if err != nil {
    if err == redis.Nil {
      return "", nil
    }
    return "", err
  }
  if len(result) != 2 {
    return "", nil
  }
  return result[1], nil
}
