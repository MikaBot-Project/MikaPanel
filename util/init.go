package util

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func LoadJsonDir[T any](rootPath string, textMap *map[string]T) {
	dirInfo, err := os.Stat(rootPath)
	if err != nil {
		fmt.Println("读取路径" + rootPath + "信息失败")
		return
	}
	if !dirInfo.IsDir() {
		fmt.Println("路径" + rootPath + "非文件夹")
		return
	}
	dir, err := os.Open(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	files, err := dir.Readdir(-1)
	err = dir.Close()
	if err != nil {
		return
	}
	for _, file := range files {
		if StringMatch(file.Name(), ".json") {
			textMessageFile, err := os.ReadFile(rootPath + "/" + file.Name())
			if err != nil {
				fmt.Println("读取" + file.Name() + "配置文件失败")
				fmt.Println(err)
				continue
			}
			var msgMap T
			err = json.Unmarshal(textMessageFile, &msgMap)
			if err != nil {
				fmt.Println("解析" + file.Name() + "配置文件失败")
				fmt.Println(err)
			}
			(*textMap)[strings.Split(file.Name(), ".json")[0]] = msgMap
		}
	}
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
