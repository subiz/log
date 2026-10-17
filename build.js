const fs = require('node:fs')
const path = require('node:path')
const ERROR_MESSAGES = require('./errors.json')

const locales = ['vi_VN', 'en_US']

// An even number of opening braces escapes a placeholder in formatString.
function placeholders(message) {
	return [...new Set([...message.matchAll(/(\{+)([A-Za-z0-9_]+)\}/g)]
		.filter(match => match[1].length % 2 === 1)
		.map(match => match[2]))].sort()
}

const problems = []
if (!ERROR_MESSAGES || typeof ERROR_MESSAGES !== 'object' || Array.isArray(ERROR_MESSAGES)) {
	throw new Error('errors.json must contain an object keyed by error code')
}
for (const [key, messages] of Object.entries(ERROR_MESSAGES)) {
	if (!/^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$/.test(key)) {
		problems.push(`${key}: error codes must use snake_case`)
	}
	if (!messages || typeof messages !== 'object' || Array.isArray(messages)) {
		problems.push(`${key}: expected an object with vi_VN and en_US messages`)
		continue
	}
	let valid = true
	for (const locale of locales) {
		if (typeof messages[locale] !== 'string' || !messages[locale].trim()) {
			problems.push(`${key}.${locale}: expected a non-empty message`)
			valid = false
		} else if (/[\uD800-\uDFFF]/u.test(messages[locale])) {
			problems.push(`${key}.${locale}: message contains an unpaired Unicode surrogate`)
			valid = false
		}
	}
	if (valid && JSON.stringify(placeholders(messages.vi_VN)) !== JSON.stringify(placeholders(messages.en_US))) {
		problems.push(`${key}: vi_VN and en_US must use the same placeholders`)
	}
}
if (problems.length) {
	throw new Error(`Invalid errors.json:\n- ${problems.join('\n- ')}`)
}

// For valid Unicode, JSON escaping produces Go-compatible quoted strings.
const quote = value => JSON.stringify(value)
let goContent = `
/* GENERATED FILE, DO NOT EDIT */
package log

type H map[string]string

var ErrorTable = map[E]H{
`

Object.keys(ERROR_MESSAGES).forEach((key) => {
	goContent += `\t${quote(key)}: H{
		"vi_VN": ${quote(ERROR_MESSAGES[key].vi_VN)},
		"en_US": ${quote(ERROR_MESSAGES[key].en_US)},
	},
`
})

goContent += `}
`

// Validate everything before replacing the generated table.
fs.writeFileSync(path.join(__dirname, 'error_table.go'), goContent)
