package router

import (
	"github.com/iesreza/gutil/log"
	"regexp"
	"strings"
)

var matches = map[string]*regexp.Regexp{
	"a":  regexp.MustCompile(`[a-zA-Z0-9\.-]+`),
	"i":  regexp.MustCompile(`[0-9]+`),
	"id": regexp.MustCompile(`[0-9]+`),
	"s":  regexp.MustCompile(`[a-zA-Z0-9\.-]+`),
}

func tokenize(input string) []token {
	var tokens []token
	parts := strings.Split(input, "/")
	for _, item := range parts {
		tk := token{}
		if len(item) > 3 {
			if item[0] == '~' {
				tk.lazy = true
			}
			if (item[0] == '[' || item[1] == '[') && item[len(item)-1] == ']' {

				meetStart := false
				meetDot := false
				varType := ""
				for i := 0; i < len(item)-1; i++ {
					if !meetStart {
						if item[i] == '[' {
							meetStart = true
						}
						continue
					}
					if item[i] == ':' {
						meetDot = true
						continue
					}
					if !meetDot {
						varType += string(item[i])
					} else {
						tk.varName += string(item[i])
					}
				}

				if val, ok := matches[varType]; ok {
					tk.match = val
				} else {
					log.Critical("Invalid variable type \"%s\" in url", item)
				}

				tk.matchType = 1
			} else {
				tk.match = item
				tk.matchType = 0
			}
		} else {
			tk.match = item
			tk.matchType = 0
		}
		tokens = append(tokens, tk)
	}

	return tokens
}

type token struct {
	match     interface{}
	varName   string
	matchType uint
	lazy      bool
}

func (t *token) isMatch(input string) bool {
	if t.matchType == 0 {
		return input == t.match.(string)
	} else if t.matchType == 1 {
		return t.match.(*regexp.Regexp).MatchString(input)
	}
	return false
}
