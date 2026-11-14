const assert = require('node:assert/strict')
const test = require('node:test')
const { format } = require('./main')
const cases = require('./testdata/format.json')

for (const { name, template, data, expected } of cases) {
	test(name, () => assert.equal(format(template, data), expected))
}
