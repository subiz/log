const ERROR_MESSAGES = require('./errors.json')
module.exports = {
	messages: ERROR_MESSAGES,
	format,
}

const ESCCHAR = '{'
const ESCCHAREND = '}'

function isAlphabet(r) {
	if ((r < 'a' || 'z' < r) && (r < 'A' || 'Z' < r) && (r < '0' || '9' < r) && r != '_' && r != ESCCHAREND) {
		return false
	}
	return true
}

function format(s, data) {
	let i = 0
	let output = ''

	while (i < s.length) {
		if (s[i] == ESCCHAR) {
			let j = i + 1
			while (j < s.length && s[j] == ESCCHAR) {
				j++
				if ((j - i) % 2 == 0) {
					output += ESCCHAR
				}
			}
			if ((j - i) % 2 != 0) {
				var param = ''
				while (j < s.length && s[j] != ESCCHAREND && isAlphabet(s[j])) {
					param += s[j]
					j++
				}
				if (param.length > 0) {
					if (!isAlphabet(s[j])) {
						output += ESCCHAR + param.substr(0, param.length - 1)
						j--
					} else {
						let v = data[param]
						if (v == undefined) {
							// no key ->
							output += ESCCHAR + param
							if (j >= s.length) {
							} else if (s[j] == ESCCHAREND) {
								output += String(ESCCHAREND)
							}
						} else {
							output += v + ''
						}
					}
				}

				if (j == s.length) {
					return output
				} else if (s[j] == ESCCHAREND) {
					j++
				}
			}
			i = j
		}
		if (i < s.length) {
			output += s[i]
		}
		i++
	}
	return output
}
