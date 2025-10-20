package util

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

func Int64ToString(in int64) string {
	return strconv.FormatInt(in, 10)
}

func StringToInt64(in string) int64 {
	out, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		log.Fatal("String To Int64:", err)
		return 0
	}
	return out
}

func StringToInt(in string) int {
	out, err := strconv.ParseInt(in, 10, 32)
	if err != nil {
		log.Fatal("String To Int:", err)
		return 0
	}
	return int(out)
}

func StringMatch(str, pattern string) bool {
	res, err := regexp.MatchString(pattern, str)
	if err != nil {
		fmt.Println()
		return false
	}
	return res
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var randomTime = int64(0)

func RandomString(length int) string {
	for randomTime == time.Now().UnixNano() {
	}
	randomTime = time.Now().UnixNano()
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func ArrayFastDelete[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	// 用最后一个元素覆盖要删除的元素
	slice[index] = slice[len(slice)-1]
	// 返回去掉最后一个元素的切片
	return slice[:len(slice)-1]
}
