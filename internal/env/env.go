package env

import (
	"os"
	"strconv"
)

// GetString возвращает значение переменной окружения или фоллбэк, если переменная не установлена.
func GetString(key string, fallback string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return res
}

// GetInt возвращает целочисленное значение переменной окружения или фоллбэк.
func GetInt(key string, fallback int) int {
	res, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	val, err := strconv.Atoi(res)
	if err != nil {
		return fallback
	}
	return val
}
