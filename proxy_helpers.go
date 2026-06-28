package main

import "net/url"

func mustParseTarget(raw string) *url.URL {
	target, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return target
}