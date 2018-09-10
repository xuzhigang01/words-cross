package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
var p = flag.Int("p", 8080, "listen port")
var dict = make(map[string]bool)

func main() {
	flag.Parse()
	initDict()
	logger.Printf("words-cross server start, listen port: %d\n", *p)

	// web server
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/genwords", genWords)
	http.HandleFunc("/buildcross", buildCross)
	logger.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(*p), nil))
}

// 加载字典
func initDict() {
	file, err := os.Open("dict.txt")
	if err != nil {
		logger.Fatalf("initDict failed: %s\n", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		} else if len(line) > 0 {
			dict[string(line)] = true
		}
	}
	logger.Printf("init dict OK, count: %d\n", len(dict))
}

// 生成单词
func genWords(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	letters := r.Form["letters"]
	if len(letters) == 0 || len(letters[0]) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("missing letters"))
		return
	}
	words := generateWords(strings.ToLower(strings.TrimSpace(letters[0])))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// 构建交叉
func buildCross(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	words := r.Form["words"]
	if len(words) == 0 || len(words[0]) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("missing words"))
		return
	}
	board, err := Build(words[0])
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

// 使用参数s中的字母组合成单词，用字典过滤掉非法单词，最后返回最长的20个单词
func generateWords(s string) []string {
	mm := make(map[string]bool)
	for _, r := range s {
		m := make(map[string]bool)
		for k, v := range mm {
			m[k] = v
		}
		m[string(r)] = true
		for k := range mm {
			len := len(k)
			for i := 0; i <= len; i++ {
				m[k[0:i]+string(r)+k[i:]] = true
			}
		}
		mm = m
	}
	words := make([]string, 0)
	for k := range mm {
		l := len(k)
		if l < 4 || l > 8 {
			continue
		}
		if _, ok := dict[k]; ok {
			words = append(words, k)
		}
	}
	if len(words) > 20 {
		sort.Sort(byLen(words))
		words = words[:20]
	}
	sort.Strings(words)
	return words
}

// 根据字符串长度排序（倒序）
type byLen []string

func (a byLen) Len() int {
	return len(a)
}

func (a byLen) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}

func (a byLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
