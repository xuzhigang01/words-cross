package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	wordsMinNum    int = 3
	wordsMaxNum    int = 20
	wordMinLength  int = 4
	boardMaxWidth  int = 18
	boardMaxHeight int = 16
)

// Board 单词交叉布局（类似棋盘）
type Board struct {
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Words  []LocWord `json:"words"`
}

// LocWord 带位置信息的单词
type LocWord struct {
	W string `json:"w"`
	X int    `json:"x"`
	Y int    `json:"y"`
	D int    `json:"d"`
}

// Build 构建交叉单词布局
func Build(words string) (*Board, error) {
	wordList := strings.Split(words, ",")
	strs := make([]string, 0, len(wordList))
	for _, s := range wordList {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			strs = append(strs, s)
		}
	}
	l := len(strs)
	if l < wordsMinNum {
		return nil, fmt.Errorf("need at lest %d words", wordsMinNum)
	} else if l > wordsMaxNum {
		return nil, fmt.Errorf("should not more than %d words", wordsMaxNum)
	}
	wordList = make([]string, 0, l)
	for _, s := range strs {
		if utf8.RuneCountInString(s) < wordMinLength {
			return nil, fmt.Errorf("invalid word: %s, word should not less than %d letters", s, wordMinLength)
		}
		wordList = append(wordList, strings.ToUpper(s))
	}
	t := time.Now()

	cross := NewWordCross(wordList, boardMaxWidth, boardMaxHeight)
	if cross.Build() {
		logger.Printf("WordsCross.Build OK, cost %.3fs\n%v\n", time.Since(t).Seconds(), cross)

		w, h := cross.GetSize()
		wps := cross.GetWordPosList()

		lws := make([]LocWord, 0, len(wps))
		for _, wp := range wps {
			lw := LocWord{W: wp.W, X: wp.X, Y: wp.Y, D: int(wp.D)}
			lws = append(lws, lw)
		}

		return &Board{w, h, lws}, nil
	}

	logger.Printf("WordsCross.Build failed, cost %.3fs\n", time.Since(t).Seconds())
	return nil, errors.New("WordsCross.Build failed")
}
