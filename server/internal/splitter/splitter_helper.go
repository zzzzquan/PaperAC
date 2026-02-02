package splitter

// hasSpecialCharacters 检查是否包含特殊字符
func hasSpecialCharacters(s string) bool {
	return specialCharRe.MatchString(s)
}
