const fs = require('node:fs')
const ERROR_MESSAGES = require('./errors.json')

let goContent = `
/* GENERATED FILE, DO NOT EDIT */
package log

type H map[string]string

var ErrorTable = map[E]H{
`

Object.keys(ERROR_MESSAGES).map((key) => {
	goContent += `	"${key}": H{
		"vi_VN": "${ERROR_MESSAGES[key].vi_VN}",
		"en_US": "${ERROR_MESSAGES[key].en_US}",
	},
`
})

goContent += `}
`

fs.writeFileSync('./error_table.go', goContent)
