package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[seededRand.Intn(len(letters))]
	}
	return string(b)
}

func Md5String(str string) string {
	w := md5.New()
	_, _ = io.WriteString(w, str)
	return hex.EncodeToString(w.Sum(nil))
}

func SaveAppStartTime(dir string) error {
	t := fmt.Sprintf("%v\n", time.Now())
	err := os.MkdirAll(filepath.Dir(dir), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "restart.txt"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(t); err != nil {
		return err
	}
	return nil
}
