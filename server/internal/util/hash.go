package util

// 验证码哈希工具，避免明文存储。

import (
  "crypto/sha256"
  "encoding/hex"
)

func HashCode(code string, salt string) string {
  hasher := sha256.New()
  hasher.Write([]byte(code))
  hasher.Write([]byte(salt))
  return hex.EncodeToString(hasher.Sum(nil))
}
