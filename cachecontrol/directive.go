package cachecontrol

import (
	"strconv"
	"strings"
	"time"
)

type CacheControl struct {
	MaxAge       *time.Duration
	SMaxAge      *time.Duration
	StaleIfError *time.Duration
	Immutable    bool
}

func isSeparator(c byte) bool {
	switch c {
	case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}', ' ', '\t':
		return true
	}
	return false
}

func isChar(c byte) bool { return c <= 127 }

func isToken(c byte) bool {
	return isChar(c) && !isSeparator(c)
}

func (cc *CacheControl) addToken(key string) {
	switch key {
	case "immutable":
		cc.Immutable = true
	}
}

func (cc *CacheControl) addPair(key string, value string) {
	switch key {
	case "max-age":
		maxAge, err := valueAsDuration(value)
		if err == nil {
			cc.MaxAge = maxAge
		}
	case "s-maxage":
		sMaxAge, err := valueAsDuration(value)
		if err == nil {
			cc.SMaxAge = sMaxAge
		}
	case "stale-if-error":
		staleIfError, err := valueAsDuration(value)
		if err == nil {
			cc.StaleIfError = staleIfError
		}
	}
}

func valueAsDuration(value string) (*time.Duration, error) {
	num, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	duration := time.Duration(num) * time.Second
	return &duration, nil
}

func takeWhileToken(str *string, i int) int {
	for i < len(*str) {
		if !isToken((*str)[i]) {
			break
		}
		i++
	}
	return i
}

func Parse(str string) (cc CacheControl) {
	i := 0
	for i < len(str) {
		if str[i] == ' ' || str[i] == ',' {
			i++
			continue
		}

		directiveEnd := takeWhileToken(&str, i+1)
		directive := strings.ToLower(str[i:directiveEnd])

		if directiveEnd+1 < len(str) && str[directiveEnd] == '=' {
			valueStart := directiveEnd + 1
			valueEnd := takeWhileToken(&str, valueStart)
			cc.addPair(directive, str[valueStart:valueEnd])
			i = valueEnd
		} else {
			cc.addToken(directive)
			i = directiveEnd
		}
	}

	return cc
}
