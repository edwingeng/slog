package slog

import (
	"regexp"
	"strings"
	"unicode"
)

type MessageFinder Scavenger

func (mf *MessageFinder) FindString(str string) (string, int, bool) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	if str != "" {
		for i, e := range mf.entries {
			if strings.Contains(e.Message, str) {
				return e.Message, i, true
			}
		}
	} else {
		for i, e := range mf.entries {
			if e.Message == "" {
				return "", i, true
			}
		}
	}
	return "", 0, false
}

func (mf *MessageFinder) FindUniqueString(str string) (string, int, bool) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	var r string
	var n int
	if str != "" {
		for i, e := range mf.entries {
			if strings.Contains(e.Message, str) {
				r = e.Message
				switch n++; n {
				case 1:
				default:
					return r, i, false
				}
			}
		}
	} else {
		for i, e := range mf.entries {
			if e.Message == "" {
				switch n++; n {
				case 1:
				default:
					return r, i, false
				}
			}
		}
	}
	return r, 0, n == 1
}

func (mf *MessageFinder) FindStringSequence(a []string) (int, bool) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	var j int
	for i := 0; i < len(mf.entries); i++ {
		if a[j] != "" {
			if strings.Contains(mf.entries[i].Message, a[j]) {
				if j++; j >= len(a) {
					break
				}
			}
		} else {
			if mf.entries[i].Message == "" {
				if j++; j >= len(a) {
					break
				}
			}
		}
	}
	return j, j == len(a)
}

func (mf *MessageFinder) FindRegexp(pat string) (string, int, bool) {
	if pat == "" {
		return mf.FindString("")
	}

	rex, err := regexp.Compile(pat)
	if err != nil {
		panic(err)
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	for i, e := range mf.entries {
		if rex.FindStringIndex(e.Message) != nil {
			return e.Message, i, true
		}
	}
	return "", 0, false
}

func (mf *MessageFinder) FindUniqueRegexp(pat string) (string, int, bool) {
	if pat == "" {
		return mf.FindUniqueString("")
	}

	rex, err := regexp.Compile(pat)
	if err != nil {
		panic(err)
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	var r string
	var n int
	for i, e := range mf.entries {
		if rex.FindStringIndex(e.Message) != nil {
			r = e.Message
			switch n++; n {
			case 1:
			default:
				return r, i, false
			}
		}
	}
	return r, 0, n == 1
}

func (mf *MessageFinder) FindRegexpSequence(a []string) (int, bool) {
	rexArr := make([]*regexp.Regexp, len(a))
	for i, pat := range a {
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

	var j int
	for i := 0; i < len(mf.entries); i++ {
		rex := rexArr[j]
		if rex != nil {
			if rex.FindStringIndex(mf.entries[i].Message) != nil {
				if j++; j >= len(a) {
					break
				}
			}
		} else {
			if mf.entries[i].Message == "" {
				if j++; j >= len(a) {
					break
				}
			}
		}
	}
	return j, j == len(a)
}

func (mf *MessageFinder) Find(str string) (string, int, bool) {
	if strings.HasPrefix(str, rexPrefix) {
		pat := strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
		return mf.FindRegexp(pat)
	} else {
		return mf.FindString(str)
	}
}

func (mf *MessageFinder) FindUnique(str string) (string, int, bool) {
	if strings.HasPrefix(str, rexPrefix) {
		pat := strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
		return mf.FindUniqueRegexp(pat)
	} else {
		return mf.FindUniqueString(str)
	}
}

func (mf *MessageFinder) FindSequence(a []string) (int, bool) {
	rexArr := make([]*regexp.Regexp, len(a))
	strArr := make([]string, len(a))
	for i, str := range a {
		if strings.HasPrefix(str, rexPrefix) {
			pat := strings.TrimLeftFunc(strings.TrimPrefix(str, rexPrefix), unicode.IsSpace)
			if pat != "" {
				if rex, err := regexp.Compile(pat); err != nil {
					panic(err)
				} else {
					rexArr[i] = rex
				}
			} else {
				strArr[i] = ""
			}
		} else {
			strArr[i] = str
		}
	}

	mf.mu.Lock()
	defer mf.mu.Unlock()

	var j int
	for i := 0; i < len(mf.entries); i++ {
		rex := rexArr[j]
		if rex != nil {
			if rex.FindString(mf.entries[i].Message) != "" {
				if j++; j >= len(a) {
					break
				}
			}
		} else if str := strArr[j]; str != "" {
			if strings.Contains(mf.entries[i].Message, str) {
				if j++; j >= len(a) {
					break
				}
			}
		} else if mf.entries[i].Message == "" {
			if j++; j >= len(a) {
				break
			}
		}
	}
	return j, j == len(a)
}
