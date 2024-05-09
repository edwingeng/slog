package slog

import (
	"regexp"
	"strings"
	"unicode"
)

type MessageFinder Scavenger

func (mf *MessageFinder) FindString(str string) []int {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	var ret []int
	if str != "" {
		for i, e := range mf.entries {
			if strings.Contains(e.Message, str) {
				ret = append(ret, i)
			}
		}
	} else {
		for i, e := range mf.entries {
			if e.Message == "" {
				ret = append(ret, i)
			}
		}
	}
	return ret
}

func (mf *MessageFinder) FindStringSequence(seq []string) ([]int, bool) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	var ret []int
	for i, e := range mf.entries {
		j := len(ret)
		if j >= len(seq) {
			break
		}
		if seq[j] != "" {
			if strings.Contains(e.Message, seq[j]) {
				ret = append(ret, i)
			}
		} else {
			if e.Message == "" {
				ret = append(ret, i)
			}
		}
	}

	ok := len(ret) == len(seq)
	return ret, ok
}

func (mf *MessageFinder) FindRegexp(pat string) []int {
	if pat == "" {
		return mf.FindString("")
	}

	rex, err := regexp.Compile(pat)
	if err != nil {
		panic(err)
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	var ret []int
	for i, e := range mf.entries {
		if rex.FindStringIndex(e.Message) != nil {
			ret = append(ret, i)
		}
	}
	return ret
}

func (mf *MessageFinder) FindRegexpSequence(seq []string) ([]int, bool) {
	rexArr := make([]*regexp.Regexp, len(seq))
	for i, pat := range seq {
		if pat != "" {
			if rex, err := regexp.Compile(pat); err != nil {
				panic(err)
			} else {
				rexArr[i] = rex
			}
		}
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	var ret []int
	for i, e := range mf.entries {
		j := len(ret)
		if j >= len(seq) {
			break
		}
		if rex := rexArr[j]; rex != nil {
			if rex.FindStringIndex(e.Message) != nil {
				ret = append(ret, i)
			}
		} else {
			if e.Message == "" {
				ret = append(ret, i)
			}
		}
	}

	ok := len(ret) == len(seq)
	return ret, ok
}

func (mf *MessageFinder) Find(str string) []int {
	if strings.HasPrefix(str, rexPrefix) {
		pat := strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
		return mf.FindRegexp(pat)
	} else {
		return mf.FindString(str)
	}
}

func (mf *MessageFinder) FindSequence(seq []string) ([]int, bool) {
	rexArr := make([]*regexp.Regexp, len(seq))
	strArr := make([]string, len(seq))
	var rexCount int
	for i, str := range seq {
		if strings.HasPrefix(str, rexPrefix) {
			pat := strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
			if pat != "" {
				if rex, err := regexp.Compile(pat); err != nil {
					panic(err)
				} else {
					rexArr[i] = rex
					rexCount++
				}
			} else {
				strArr[i] = ""
			}
		} else {
			strArr[i] = str
		}
	}
	if rexCount == 0 {
		return mf.FindStringSequence(seq)
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	var ret []int
	for i, e := range mf.entries {
		j := len(ret)
		if j >= len(seq) {
			break
		}
		if rex := rexArr[j]; rex != nil {
			if rex.FindStringIndex(e.Message) != nil {
				ret = append(ret, i)
			}
		} else if str := strArr[j]; str != "" {
			if strings.Contains(e.Message, str) {
				ret = append(ret, i)
			}
		} else {
			if e.Message == "" {
				ret = append(ret, i)
			}
		}
	}

	ok := len(ret) == len(seq)
	return ret, ok
}
