package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"time"
	"unicode/utf8"
)

// Direction 方向
type Direction int

// Position 位置
type Position struct {
	X int
	Y int
	D Direction
}

// WordPos 单词和位置
type WordPos struct {
	W string
	Position
}

const (
	dX Direction = iota
	dY
)

// WordsCross 单词交叉类型
type WordsCross struct {
	words []string
	maxW  int
	maxH  int
	incX  bool
	grids [][]*cell
	wps   []WordPos
}

// NewWordCross 创建WordsCross，参数：单词列表、最大宽度、最大高度
func NewWordCross(wordList []string, maxWidth, maxHeight int) *WordsCross {
	return &WordsCross{words: wordList, maxW: maxWidth, maxH: maxHeight, incX: true}
}

// Build 构建单词交叉矩阵，返回是否成功
func (o *WordsCross) Build() bool {
	o.initSize()
	numSucc := 0
	numFail := 0
	for {
		if o.buildBestCross() {
			numSucc++
			// 每个尺寸，尝试10次构建全交叉
			if !o.isAllCrossed() && numSucc < 10 {
				continue
			}
			logger.Printf("numSucc: %d\n", numSucc)
			o.slim()
			break
		} else {
			numFail++
			// 每个尺寸，尝试100次后增加矩阵尺寸
			if numFail < 100 {
				continue
			}
			if !o.increSize() {
				return false
			}
			numSucc = 0
			numFail = 0
		}
	}
	return true
}

// GetSize 获取矩阵尺寸，返回：宽、高
func (o *WordsCross) GetSize() (width, height int) {
	return len(o.grids[0]), len(o.grids)
}

// GetWordPosList 获取单词及其位置列表
func (o *WordsCross) GetWordPosList() []WordPos {
	return o.wps
}

// 初始化矩阵尺寸
func (o *WordsCross) initSize() {
	letterCount := 0 // 字母总数
	longest := 0     // 最长单词长度
	shortest := 100  // 最短单词长度
	for _, w := range o.words {
		l := len(w)
		letterCount += l
		if longest < l {
			longest = l
		}
		if shortest > l {
			shortest = l
		}
	}

	area := math.Ceil(float64(letterCount) * 1.5)
	a := math.Ceil(math.Sqrt(area))
	w := math.Max(float64(longest), a)
	h := math.Min(a, math.Ceil(area/w))
	h = math.Max(h, float64(shortest))

	width := int(math.Min(w, float64(o.maxW)))
	height := int(math.Min(h, float64(o.maxH)))

	o.initGrids(width, height)
}

// 增加矩阵尺寸
func (o *WordsCross) increSize() bool {
	w, h := o.GetSize()
	if w >= o.maxW && h >= o.maxH {
		return false
	}
	if o.incX && w < o.maxW {
		w++
	} else if h < o.maxH {
		h++
	} else {
		w++
	}
	o.incX = !o.incX
	o.initGrids(w, h)
	return true
}

// 初始化矩阵
func (o *WordsCross) initGrids(width, height int) {
	logger.Printf("initGrids: %dx%d\n", width, height)
	grids := make([][]*cell, 0, height)
	for y := 0; y < height; y++ {
		row := make([]*cell, 0, width)
		for x := 0; x < width; x++ {
			row = append(row, new(cell))
		}
		grids = append(grids, row)
	}
	o.grids = grids
}

// 清空矩阵内容
func (o *WordsCross) clearGrids() {
	for _, row := range o.grids {
		for _, c := range row {
			c.clear()
		}
	}
	o.wps = make([]WordPos, 0, len(o.words))
}

// 打乱单词顺序
func (o *WordsCross) shuffleWords(r *rand.Rand) {
	for n := len(o.words); n > 0; n-- {
		i := r.Intn(n)
		o.words[n-1], o.words[i] = o.words[i], o.words[n-1]
	}
}

// 构建最优的单词交叉布局，返回是否成功
func (o *WordsCross) buildBestCross() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	o.shuffleWords(r)
	o.clearGrids()

	n := len(o.words)

	used := make([]bool, n, n)
	pos := make([]*Position, n, n)
	scores := make([]int, n, n)

	for remaining := n; remaining > 0; remaining-- {
		// 剩下单词的最佳位置及得分
		for i := 0; i < n; i++ {
			if !used[i] {
				pos[i], scores[i] = nil, -1
				p, s := o.getBestPosition(&(o.words[i]))
				if p == nil {
					//logger.Println(o.words[i])
					//logger.Println(o)
					return false
				}
				pos[i], scores[i] = p, s
			}
		}

		// 如果最高分的位置有多个，则随机选择一个
		numbest := 0
		bestScore := -1
		for i := 0; i < n; i++ {
			if !used[i] {
				s := scores[i]
				if s == bestScore {
					numbest++
				} else if s > bestScore {
					bestScore = s
					numbest = 1
				}
			}
		}
		choose := 0
		j := 0
		if numbest > 1 {
			j = r.Intn(numbest)
		}
		for i, k := 0, 0; i < n; i++ {
			if !used[i] && scores[i] == bestScore {
				if k == j {
					choose = i
					break
				}
				k++
			}
		}

		// 放置单词到最佳位置
		o.placeWord(&(o.words[choose]), pos[choose])
		used[choose] = true

		o.wps = append(o.wps, WordPos{o.words[choose], *pos[choose]})
	}

	return true
}

// 获取最佳位置(及得分)，没有可用的位置返回nil,-1
func (o *WordsCross) getBestPosition(word *string) (*Position, int) {
	var pos *Position
	score := -1

	h := len(o.grids)
	w := len(o.grids[0])
	l := utf8.RuneCountInString(*word)

	// 遍历每个位置，找出最高分的位置
	// 横放
	endX := w - l + 1
	for y := 0; y < h; y++ {
		for x := 0; x < endX; x++ {
			p := &Position{x, y, dX}
			if o.checkPosition(p, word) {
				s := o.scorePosition(p, word)
				if s > score {
					pos = p
					score = s
				}
			}
		}
	}
	// 竖置
	endY := h - l + 1
	for y := 0; y < endY; y++ {
		for x := 0; x < w; x++ {
			p := &Position{x, y, dY}
			if o.checkPosition(p, word) {
				s := o.scorePosition(p, word)
				if s > score {
					pos = p
					score = s
				}
			}
		}
	}
	return pos, score
}

// 检查位置是否可用
func (o *WordsCross) checkPosition(p *Position, word *string) bool {
	l := utf8.RuneCountInString(*word)
	return o.noPrePost(p, l) && o.noInner(p, l) && o.noAdjacent(p, l) && o.crossMatch(p, word)
}

// 头尾（横向：左右，竖向：上下）没有字母
func (o *WordsCross) noPrePost(p *Position, l int) bool {
	y, x := p.Y, p.X
	if p.D == dX {
		row := o.grids[y]
		if x > 0 && row[x-1].content > 0 {
			return false
		}
		if x+l < len(o.grids[0]) && row[x+l].content > 0 {
			return false
		}
	} else {
		if y > 0 && o.grids[y-1][x].content > 0 {
			return false
		}
		if y+l < len(o.grids) && o.grids[y+l][x].content > 0 {
			return false
		}
	}
	return true
}

// 该区域内，没有单词
func (o *WordsCross) noInner(p *Position, l int) bool {
	if p.D == dX {
		row := o.grids[p.Y]
		count := 0
		for x, end := p.X, p.X+l; x < end; x++ {
			if row[x].content > 0 {
				count++
				if count > 1 {
					return false
				}
			} else {
				count = 0
			}
		}
	} else {
		x := p.X
		count := 0
		for y, end := p.Y, p.Y+l; y < end; y++ {
			if o.grids[y][x].content > 0 {
				count++
				if count > 1 {
					return false
				}
			} else {
				count = 0
			}
		}
	}
	return true
}

// 相邻（横向：上下，竖向：左右）没有单词
func (o *WordsCross) noAdjacent(p *Position, l int) bool {
	if p.D == dX {
		y := p.Y
		begin := p.X
		if begin > 0 {
			begin--
		}
		end := p.X + l
		if end < len(o.grids[0]) {
			end++
		}
		if y > 0 {
			row := o.grids[y-1]
			count := 0
			for i := begin; i < end; i++ {
				c := row[i]
				if c.content > 0 {
					if c.end {
						return false
					}
					count++
					if count > 1 {
						return false
					}
				} else {
					count = 0
				}
			}
		}
		if y+1 < len(o.grids) {
			row := o.grids[y+1]
			count := 0
			for i := begin; i < end; i++ {
				c := row[i]
				if c.content > 0 {
					if c.start {
						return false
					}
					count++
					if count > 1 {
						return false
					}
				} else {
					count = 0
				}
			}
		}
	} else {
		begin := p.Y
		if begin > 0 {
			begin--
		}
		end := p.Y + l
		if end < len(o.grids) {
			end++
		}
		if p.X > 0 {
			x := p.X - 1
			count := 0
			for i := begin; i < end; i++ {
				c := o.grids[i][x]
				if c.content > 0 {
					if c.end {
						return false
					}
					count++
					if count > 1 {
						return false
					}
				} else {
					count = 0
				}
			}
		}
		if p.X+1 < len(o.grids[0]) {
			x := p.X + 1
			count := 0
			for i := begin; i < end; i++ {
				c := o.grids[i][x]
				if c.content > 0 {
					if c.start {
						return false
					}
					count++
					if count > 1 {
						return false
					}
				} else {
					count = 0
				}
			}
		}
	}
	return true
}

// 字母是否匹配。如有交叉，检查交叉字母是否匹配
func (o *WordsCross) crossMatch(p *Position, word *string) bool {
	if p.D == dX {
		row := o.grids[p.Y]
		x := p.X
		for _, r := range *word {
			c := row[x].content
			if c > 0 && c != r {
				return false
			}
			x++
		}
	} else {
		y := p.Y
		x := p.X
		for _, r := range *word {
			c := o.grids[y][x].content
			if c > 0 && c != r {
				return false
			}
			y++
		}
	}
	return true
}

// 给位置打分，返回分数
func (o *WordsCross) scorePosition(p *Position, word *string) int {
	h := len(o.grids)
	w := len(o.grids[0])
	l := utf8.RuneCountInString(*word)

	x := p.X
	y := p.Y

	// 单词越长、离中心越近，得分越高；纵向优先；靠边扣分
	score := 0
	if p.D == dX {
		hh := y + 1
		if hh*2 > h {
			hh = h - hh
		}
		score += l * hh / 2
		row := o.grids[y]
		for i, end := x, x+l; i < end; i++ {
			if row[i].content > 0 {
				score += 3
			}
		}
		if y == 0 || y == h-1 {
			score -= l / 2
		}
	} else {
		hh := x + 1
		if hh*2 > w {
			hh = w - hh
		}
		score += l * hh / 2
		for i, end := y, y+l; i < end; i++ {
			if o.grids[i][x].content > 0 {
				score += 4
			}
		}
		if x == 0 || x == w-1 {
			score -= l / 2
		}
	}
	return score
}

// 放置单词
func (o *WordsCross) placeWord(word *string, p *Position) {
	if p.D == dX {
		row := o.grids[p.Y]
		x := p.X
		row[x].start = true
		for _, r := range *word {
			c := row[x]
			if c.content > 0 {
				c.cross = true
			} else {
				c.content = r
			}
			x++
		}
		row[x-1].end = true
	} else {
		y := p.Y
		x := p.X
		o.grids[y][x].start = true
		for _, r := range *word {
			c := o.grids[y][x]
			if c.content > 0 {
				c.cross = true
			} else {
				c.content = r
			}
			y++
		}
		o.grids[y-1][x].end = true
	}
}

// 是否全交叉（每个单词都有交叉点）
func (o *WordsCross) isAllCrossed() bool {
	for _, wp := range o.wps {
		l := utf8.RuneCountInString(wp.W)
		cross := false
		if wp.D == dX {
			row := o.grids[wp.Y]
			for x, end := wp.X, wp.X+1; x < end; x++ {
				if row[x].cross {
					cross = true
					break
				}
			}
		} else {
			x := wp.X
			for y, end := wp.Y, wp.Y+l; y < end; y++ {
				if o.grids[y][x].cross {
					cross = true
					break
				}
			}
		}
		if !cross {
			return false
		}
	}
	return true
}

// 瘦身：删除周边的空行或空列
func (o *WordsCross) slim() {
	emptyFirstRow := true
	for _, cell := range o.grids[0] {
		if cell.content > 0 {
			emptyFirstRow = false
			break
		}
	}
	if emptyFirstRow {
		o.grids = o.grids[1:]
	}

	emptyLastRow := true
	for _, cell := range o.grids[len(o.grids)-1] {
		if cell.content > 0 {
			emptyLastRow = false
			break
		}
	}
	if emptyLastRow {
		o.grids = o.grids[:len(o.grids)-1]
	}

	emptyFirstColumn := true
	for _, row := range o.grids {
		if row[0].content > 0 {
			emptyFirstColumn = false
			break
		}
	}
	if emptyFirstColumn {
		h := len(o.grids)
		for y := 0; y < h; y++ {
			o.grids[y] = o.grids[y][1:]
		}
	}

	emptyLastColumn := true
	w := len(o.grids[0])
	for _, row := range o.grids {
		if row[w-1].content > 0 {
			emptyLastColumn = false
			break
		}
	}
	if emptyLastColumn {
		h := len(o.grids)
		for y := 0; y < h; y++ {
			o.grids[y] = o.grids[y][:w-1]
		}
	}

	// 继续瘦身
	if emptyFirstRow || emptyLastRow || emptyFirstColumn || emptyLastColumn {
		o.slim()
	}
}

func (o *WordsCross) String() string {
	w, h := o.GetSize()
	var buf bytes.Buffer
	buf.WriteString("WordsCross: ")
	fmt.Fprintf(&buf, "%dx%d\n\n", w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.WriteRune(' ')
			r := o.grids[y][x].content
			if r > 0 {
				buf.WriteString(string(r))
			} else {
				buf.WriteRune('_')
			}
		}
		buf.WriteRune('\n')
	}
	return buf.String()
}

// 单元格
type cell struct {
	start   bool
	end     bool
	cross   bool
	content rune
}

func (o *cell) clear() {
	o.start = false
	o.end = false
	o.cross = false
	o.content = 0
}
