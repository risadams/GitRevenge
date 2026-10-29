import { describe, expect, it } from 'vitest'
import { flatten, formatOf, parseStructured, visibleNodes } from './structured'

describe('parseStructured', () => {
  it('reads JSON, YAML and TOML', () => {
    expect(parseStructured('a.json', '{"x":[1,2]}')).toEqual({ ok: true, value: { x: [1, 2] } })
    expect(parseStructured('a.yml', 'x:\n  - 1\n  - two\n')).toEqual({ ok: true, value: { x: [1, 'two'] } })
    expect(parseStructured('a.toml', 'title = "t"\n[owner]\nname = "n"\n')).toEqual({ ok: true, value: { title: 't', owner: { name: 'n' } } })
  })

  it('lists the documents of a multi-document YAML file', () => {
    expect(parseStructured('a.yaml', 'a: 1\n---\nb: 2\n')).toEqual({ ok: true, value: [{ a: 1 }, { b: 2 }] })
  })

  it('reports parse errors', () => {
    for (const [p, t] of [['a.json', '{"x":'], ['a.toml', 'x = '], ['a.yaml', 'a: [1, 2']]) {
      const res = parseStructured(p, t)
      expect(res.ok, p).toBe(false)
      if (!res.ok) expect(res.error.length).toBeGreaterThan(0)
    }
  })

  it('knows the formats by extension', () => {
    expect(['x.json', 'x.YAML', 'x.toml', 'x.md'].map(formatOf)).toEqual(['json', 'yaml', 'toml', null])
  })
})

describe('tree', () => {
  const nodes = flatten({ name: 'app', deps: { yaml: '2.9.1', marked: '18.1.0' }, list: [true, null] })

  it('flattens in document order with sizes for containers', () => {
    expect(nodes.map((n) => `${n.id}=${n.summary}`)).toEqual([
      '$={3}',
      '$.name="app"',
      '$.deps={2}',
      '$.deps.yaml="2.9.1"',
      '$.deps.marked="18.1.0"',
      '$.list=[2]',
      '$.list[0]=true',
      '$.list[1]=null',
    ])
  })

  it('hides the inside of collapsed nodes', () => {
    expect(visibleNodes(nodes, new Set(['$.deps']), '').map((n) => n.id)).toEqual(['$', '$.name', '$.deps', '$.list', '$.list[0]', '$.list[1]'])
  })

  it('shows matches and their ancestors, even when collapsed', () => {
    expect(visibleNodes(nodes, new Set(['$.deps']), 'MARKED').map((n) => n.id)).toEqual(['$', '$.deps', '$.deps.marked'])
    expect(visibleNodes(nodes, new Set(), '2.9').map((n) => n.id)).toEqual(['$', '$.deps', '$.deps.yaml'])
  })
})
