package main

import (
	"bufio"
	"flag"
	"github.com/hoisie/web"
	"github.com/mattn/go-lingr"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var re = regexp.MustCompile(`^(\S+)\s+#(しりとり|siritori)\s*$`)
var re2 = regexp.MustCompile(`^\s*#(しりとり|siritori)!\s*$`)

var cwd string

func defaultAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":80"
	}
	return ":" + port
}

var addr = flag.String("addr", defaultAddr(), "server address")

var upper = strings.NewReplacer(
	"ぁ", "あ",
	"ぃ", "い",
	"ぅ", "う",
	"ぇ", "え",
	"ぉ", "お",
	"ゃ", "や",
	"ゅ", "ゆ",
	"ょ", "よ",
)

func kana2hira(s string) string {
	return strings.Map(func(r rune) rune {
		if 0x30A1 <= r && r <= 0x30F6 {
			return r - 0x0060
		}
		return r
	}, s)
}

func hira2kana(s string) string {
	return strings.Map(func(r rune) rune {
		if 0x3041 <= r && r <= 0x3096 {
			return r + 0x0060
		}
		return r
	}, s)
}

func search(text string) string {
	rs := []rune(text)
	r := rs[len(rs)-1]

	f, err := os.Open(filepath.Join(cwd, "dict.txt"))
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := bufio.NewReader(f)

	words := []string{}
	for {
		b, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		line := string(b)
		if ([]rune(line))[0] == r {
			words = append(words, line)
		}
	}
	if len(words) == 0 {
		return ""
	}
	return words[rand.Int()%len(words)]
}

func shiritori(text string) string {
	text = strings.Replace(text, "ー", "", -1)
	if rand.Int()%2 == 0 {
		text = hira2kana(text)
	} else {
		text = kana2hira(text)
	}
	return search(text)
}

func handleText(text string) string {
	rs := []rune(strings.TrimSpace(text))
	if len(rs) == 0 {
		return "しばくぞ"
	}
	if rs[len(rs)-1] == 'ん' || rs[len(rs)-1] == 'ン' {
		return "勝った（笑）"
	}
	s := shiritori(text)
	if s == "" {
		return "わかりません"
	}
	rs = []rune(s)
	if rs[len(rs)-1] == 'ん' || rs[len(rs)-1] == 'ン' {
		s += "\nあっ..."
	}
	return s
}

func main() {
	flag.Parse()

	cwd = filepath.Dir(os.Args[0])

	rand.Seed(time.Now().UnixNano())

	siritoriMode := map[string]bool{}

	web.Post("/", func(ctx *web.Context) string {
		status, err := lingr.DecodeStatus(ctx.Request.Body)
		if err != nil {
			ctx.Abort(500, err.Error())
			return err.Error()
		}
		for _, event := range status.Events {
			if message := event.Message; message != nil {
				text := message.Text
				current, _ := siritoriMode[message.Room]
				if re2.MatchString(text) {
					current = !current
					siritoriMode[message.Room] = current
					if current {
						return "しりとりモード オン"
					}
					return "しりとりモード オフ"
				} else if current {
					return handleText(text)
				} else if re.MatchString(text) {
					text = re.FindStringSubmatch(text)[1]
					return handleText(text)
				}
			}
		}
		return ""
	})
	web.Run(*addr)
}
