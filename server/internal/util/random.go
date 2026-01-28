package util

// 随机码生成工具。

import (
  "crypto/rand"
  "fmt"
  "math/big"
)

func GenerateNumericCode(length int) (string, error) {
  if length <= 0 {
    return "", fmt.Errorf("长度无效")
  }
  max := big.NewInt(10)
  code := make([]byte, length)
  for i := 0; i < length; i++ {
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
      return "", err
    }
    code[i] = byte('0' + n.Int64())
  }
  return string(code), nil
}
