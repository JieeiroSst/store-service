import { Resvg } from '@resvg/resvg-js'
import { readFileSync, writeFileSync } from 'fs'

const [, , input, output, size] = process.argv
const svg = readFileSync(input)
const png = new Resvg(svg, { fitTo: { mode: 'width', value: Number(size) }, background: 'rgba(0,0,0,0)' }).render().asPng()
writeFileSync(output, png)
