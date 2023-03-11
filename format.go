package log

import "fmt"

const ESCCHAR = '{'
const ESCCHAREND = '}'

func isAlphabet(r rune) bool {
	if (r < 'a' || 'z' < r) && (r < 'A' || 'Z' < r) && (r < '0' || '9' < r) && r != '_' && r != ESCCHAREND {
		return false
	}
	return true
}

func copyr(s string) []rune {
	rs := make([]rune, 0, len(s))
	for _, r := range s {
		rs = append(rs, r)
	}
	return rs
}

// Format format s using map m
// Keys of map m can be anything but must not contains spaces
func formatString(us string, data map[string]interface{}) string {
	s := copyr(us)
	i := 0
	output := ""
	for i < len(s) {
		if s[i] == ESCCHAR {
			j := i + 1
			for j < len(s) && s[j] == ESCCHAR {
				j++
				if (j-i)%2 == 0 {
					output += string(ESCCHAR)
				}
			}
			if (j-i)%2 != 0 {
				param := ""
				for j < len(s) && s[j] != ESCCHAREND && isAlphabet(s[j]) {
					param += string(s[j])
					j++
				}
				if len(param) > 0 {
					if !isAlphabet(s[j]) {
						output = output + string(ESCCHAR) + param[0:len(param)-1]
						j--
					} else {
						v, ok := data[param]
						if !ok { // no key ->
							output = output + string(ESCCHAR) + param
							if j >= len(s) {
							} else if s[j] == ESCCHAREND {
								output += string(ESCCHAREND)
							}
						} else {
							output += fmt.Sprintf("%v", v)
						}
					}
				}

				if j == len(s) {
					return output
				} else if s[j] == ESCCHAREND {
					j++
				}
			}
			i = j
		}
		if i < len(s) {
			output += string(s[i])
		}
		i++
	}
	return output
}
