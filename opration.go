package main

import (
	"log"
	"regexp"
	"strconv"
)

//Opration 操作
type Opration = func(blance float64) float64

//OpMatcher 操作匹配器
type OpMatcher interface {
	Match(content string) Opration
}

//PatternOpMatcher 正则配置的操作匹配器
type PatternOpMatcher struct {
	Pattern string
	Op      string
	regex   *regexp.Regexp
}

//Match 匹配内容返回操作函数
func (matcher PatternOpMatcher) Match(content string) Opration {
	if matcher.regex == nil {
		matcher.regex = regexp.MustCompile(matcher.Pattern)
	}
	match := matcher.regex.FindStringSubmatch(content)
	if match == nil {
		return nil
	}
	mount, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		log.Println(err)
		return nil
	}
	return func(balance float64) float64 {
		switch matcher.Op {
		case "+":
			return balance + mount
		case "-":
			return balance - mount
		}
		log.Println("不支持的操作符：", matcher.Op)
		return balance
	}
}

//GetMatchers 返回匹配器
func GetMatchers() []OpMatcher {
	if globalMatchers == nil {
		globalMatchers = make([]OpMatcher, 0)
		for _, matcherConfig := range GlobalConfig.Patterns {
			globalMatchers = append(globalMatchers, PatternOpMatcher{
				Pattern: matcherConfig.Pattern,
				Op:      matcherConfig.Op,
			})
		}
	}
	return globalMatchers
}
