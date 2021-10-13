package main

import (
	"fmt"
	"strconv"
)

func main() {
	var s string
	fmt.Scanln(&s)

	var tmps, tmpn, res []byte
	var num int

	n := len(s)

	for i := 0; i < n; {
		if 'a' <= s[i] && s[i] <= 'z' {
			tmps = append(tmps, s[i])
			i++
		} else {
			if num > 0 {
				for j := 0; j < num; j++ {
					res = append(res, tmps...)
				}
			} else {
				res = append(res, tmps...)
			}
			tmps = []byte{}
			tmpn = []byte{}
			num = 0

			for '1' <= s[i] && s[i] <= '9' {
				tmpn = append(tmpn, s[i])
				i++
			}
			num, _ = strconv.Atoi(string(tmpn))
		}
	}

	if len(tmps) != 0 && num != 0 {
		for j := 0; j < num; j++ {
			res = append(res, tmps...)
		}
	}

	fmt.Println(string(res))
}
