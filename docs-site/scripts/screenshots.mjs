// Takes the docs' screenshots and animations from the running app.
//
// The app runs as `wails dev`, which serves the real frontend and Go backend to a browser at
// http://localhost:34115. This drives Chrome against it with Puppeteer, resets the demo
// repository (scripts/demo-repo.sh) before each scene, and writes WebP files to docs/img/:
// stills in a light and a dark version, and animations as animated WebP.
//
// To run it (Windows; the paths are what the screenshots show):
//
//   1. Map a drive to an empty folder, so paths in the shots read T:\tidewater:
//        subst T: C:\some\empty\folder
//   2. Point the app's settings at a scratch folder, so your own repositories stay out of
//      the shots, and start it:
//        $env:APPDATA = "C:\some\scratch\appdata"; wails dev -browser=false
//   3. From docs-site/:
//        node scripts/screenshots.mjs [scene ...]
//
// It needs Chrome, Git Bash, ffmpeg and ImageMagick (magick). CHROME, BASH_EXE, APP_URL and
// DEMO override where Chrome and bash are, where the app is served, and the demo folder.
import { execFileSync } from 'node:child_process'
import { mkdirSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import puppeteer from 'puppeteer-core'

const here = path.dirname(fileURLToPath(import.meta.url))
const out = path.resolve(here, '../../docs/img')
const work = path.join(tmpdir(), 'gitrevenge-shots')
const APP = process.env.APP_URL ?? 'http://localhost:34115/'
const DEMO = process.env.DEMO ?? 'T:\\'
const CHROME = process.env.CHROME ?? 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe'
// Git's bash, not WSL's, which is also on PATH as bash.exe.
const BASH = process.env.BASH_EXE ?? 'C:\\Program Files\\Git\\bin\\bash.exe'
const REPO = path.win32.join(DEMO, 'tidewater')
const WIDTH = 1280
const HEIGHT = 800

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

/** Runs a program, and on failure says what it printed. */
function run(file, args) {
  try {
    return execFileSync(file, args, { stdio: 'pipe', encoding: 'utf8' })
  } catch (e) {
    throw new Error(`${path.basename(file)} ${args.join(' ')}: ${e.stderr || e.message}`)
  }
}

// ── The demo repository and the app's session ─────────────────────────────────────────

/** Closes every tab, rebuilds the demo repositories, and opens the ones named. With clean,
 * tidewater's uncommitted changes are dropped first; with branch, that branch is checked out. */
async function reset(page, open = ['tidewater'], { clean = false, branch } = {}) {
  await page.evaluate(async () => {
    const st = await window.go.app.App.RestoreSession()
    for (const t of st?.tabs ?? []) await window.go.app.App.CloseRepo(t.path)
    await window.go.app.App.SetActiveRepo('')
  })
  // Git processes the app started can hold the folder for a moment after the tab closes.
  for (let tries = 1; ; tries++) {
    try {
      run(BASH, [path.join(here, 'demo-repo.sh'), DEMO.replace(/\\/g, '/')])
      break
    } catch (e) {
      if (tries === 8 || !/busy|denied/i.test(e.message)) throw e
      await sleep(1500)
    }
  }
  if (clean) run('git', ['-C', REPO, 'reset', '-q', '--hard'])
  if (branch) run('git', ['-C', REPO, 'checkout', '-q', branch])
  // Only the demo repositories belong on the start page.
  await page.evaluate(async (demo) => {
    for (const r of (await window.go.app.App.RecentRepos()) ?? []) {
      if (!r.path.toLowerCase().startsWith(demo.toLowerCase())) await window.go.app.App.RemoveRecentRepo(r.path)
    }
  }, DEMO)
  for (const name of open) {
    await page.evaluate((p) => window.go.app.App.OpenRepo(p), path.win32.join(DEMO, name))
  }
  await page.evaluate((p) => window.go.app.App.SetActiveRepo(p), open.length ? path.win32.join(DEMO, open[0]) : '')
  await page.reload({ waitUntil: 'networkidle0' })
  await overlay(page)
  if (open.length) await page.waitForSelector('.scroller')
  await sleep(800)
}

// ── What a viewer sees in an animation: a cursor, clicks and keys ─────────────────────

/** Adds a pointer that follows the mouse, a ring on each click, and a caption for keys. */
async function overlay(page) {
  await page.evaluate(() => {
    if (document.getElementById('demo-cursor')) return
    const css = document.createElement('style')
    css.textContent = `
      #demo-cursor { position: fixed; z-index: 99999; pointer-events: none; width: 22px; height: 22px;
        left: -40px; top: -40px; transition: transform 80ms; }
      #demo-cursor.down { transform: scale(0.85); }
      .demo-ring { position: fixed; z-index: 99998; pointer-events: none; width: 36px; height: 36px;
        margin: -18px 0 0 -18px; border-radius: 50%; border: 3px solid #cb0162; opacity: 0.9;
        animation: demo-ring 450ms ease-out forwards; }
      @keyframes demo-ring { to { transform: scale(1.8); opacity: 0; } }
      #demo-keys { position: fixed; z-index: 99999; left: 50%; bottom: 48px; transform: translateX(-50%);
        display: none; gap: 6px; padding: 10px 14px; border-radius: 10px; background: rgba(26, 22, 20, 0.88);
        box-shadow: 0 8px 30px rgba(0, 0, 0, 0.3); font: 600 18px 'JetBrains Mono', ui-monospace, monospace; }
      #demo-keys kbd { color: #f4ecdd; background: #3a322e; border: 1px solid #5a4f48; border-bottom-width: 3px;
        border-radius: 6px; padding: 3px 10px; font: inherit; }`
    document.head.append(css)
    const cursor = document.createElement('div')
    cursor.id = 'demo-cursor'
    cursor.innerHTML =
      '<svg viewBox="0 0 22 22" width="22" height="22"><path d="M3 2 L3 18 L7.5 13.8 L10.6 20.6 L13.4 19.4 L10.3 12.7 L16.5 12.4 Z" fill="#1a1614" stroke="#fff" stroke-width="1.5" stroke-linejoin="round"/></svg>'
    const keys = document.createElement('div')
    keys.id = 'demo-keys'
    document.body.append(cursor, keys)
    addEventListener('mousemove', (e) => {
      cursor.style.left = `${e.clientX - 3}px`
      cursor.style.top = `${e.clientY - 2}px`
    }, true)
    addEventListener('mousedown', (e) => {
      cursor.classList.add('down')
      const ring = document.createElement('div')
      ring.className = 'demo-ring'
      ring.style.left = `${e.clientX}px`
      ring.style.top = `${e.clientY}px`
      document.body.append(ring)
      setTimeout(() => ring.remove(), 500)
    }, true)
    addEventListener('mouseup', () => cursor.classList.remove('down'), true)
  })
}

/** Hides the cursor, for stills. */
const hideCursor = (page) => page.evaluate(() => document.getElementById('demo-cursor')?.style.setProperty('left', '-40px'))

/** Shows keys being pressed, such as "Ctrl+K", while pressing them. */
async function keys(page, combo, { hold = 900 } = {}) {
  const parts = combo.split('+')
  await page.evaluate((parts) => {
    const k = document.getElementById('demo-keys')
    k.innerHTML = parts.map((p) => `<kbd>${p}</kbd>`).join('')
    k.style.display = 'flex'
  }, parts)
  const names = parts.map((p) => ({ Ctrl: 'Control', Esc: 'Escape' })[p] ?? p)
  for (const n of names.slice(0, -1)) await page.keyboard.down(n)
  await page.keyboard.press(names.at(-1))
  for (const n of names.slice(0, -1).reverse()) await page.keyboard.up(n)
  await sleep(hold)
  await page.evaluate(() => (document.getElementById('demo-keys').style.display = 'none'))
}

let mouse = { x: WIDTH / 2, y: HEIGHT / 2 }

/** Moves the pointer to x, y in a smooth, human-looking path. */
async function glide(page, x, y, ms = 450) {
  const steps = Math.max(8, Math.round(ms / 16))
  const from = { ...mouse }
  for (let i = 1; i <= steps; i++) {
    const t = i / steps
    const e = t < 0.5 ? 2 * t * t : 1 - (-2 * t + 2) ** 2 / 2
    await page.mouse.move(from.x + (x - from.x) * e, from.y + (y - from.y) * e)
    await sleep(ms / steps)
  }
  mouse = { x, y }
}

/** The middle of the first element matching selector. */
async function centre(page, selector) {
  const el = await page.waitForSelector(selector, { visible: true })
  const box = await el.boundingBox()
  return { x: box.x + box.width / 2, y: box.y + box.height / 2 }
}

/** Glides to an element and clicks it. */
async function click(page, selector, { button = 'left', pause = 350, offset } = {}) {
  const c = await centre(page, selector)
  const x = c.x + (offset?.x ?? 0)
  const y = c.y + (offset?.y ?? 0)
  await glide(page, x, y)
  await page.mouse.click(x, y, { button })
  await sleep(pause)
}

/** Glides to a point and clicks it. */
async function clickAt(page, x, y, { button = 'left', pause = 350, count = 1 } = {}) {
  await glide(page, x, y)
  await page.mouse.click(x, y, { button, count })
  await sleep(pause)
}

/** Types like a person: one key at a time. */
async function type(page, text, delay = 45) {
  await page.keyboard.type(text, { delay })
}

/** A button, link or menu item whose text contains text. */
const byText = (text, tag = '*') =>
  `::-p-xpath(//${tag}[(self::button or self::a or @role='menuitem' or self::summary) and contains(normalize-space(.), ${JSON.stringify(text)})])`

// ── The commit graph ──────────────────────────────────────────────────────────────────

/** Where row i of the graph is on screen (its middle), counting from the top row shown. */
async function row(page, i) {
  return page.evaluate((i) => {
    const s = document.querySelector('.scroller')
    const r = s.getBoundingClientRect()
    const h = document.documentElement.dataset.density === 'compact' ? 30 : 52
    return { x: r.left + Math.min(r.width / 2, 360), y: r.top + (i + 0.5) * h, left: r.left }
  }, i)
}

/** Selects the commit whose subject is given, by searching for it. */
async function selectCommit(page, subject) {
  const box = await page.waitForSelector('::-p-aria(Search commits)')
  await box.click({ count: 3 })
  await page.keyboard.type(subject)
  await page.keyboard.press('Enter')
  await sleep(700)
  await page.keyboard.press('Escape')
  await sleep(300)
}

// ── Output ────────────────────────────────────────────────────────────────────────────

async function setTheme(page, theme) {
  await page.emulateMediaFeatures([{ name: 'prefers-color-scheme', value: theme }])
  await sleep(250)
}

/** Saves the page (or clip, a box in CSS pixels) as docs/img/<name>-<theme>.webp. */
async function still(page, name, theme, clip) {
  // Out of the way, so nothing is left looking hovered.
  await page.mouse.move(WIDTH - 2, HEIGHT - 2)
  await hideCursor(page)
  await sleep(150)
  const png = path.join(work, `${name}-${theme}.png`)
  await page.screenshot({ path: png, clip })
  execFileSync('magick', [png, '-quality', '90', '-define', 'webp:method=6', path.join(out, `${name}-${theme}.webp`)])
  console.log(`  ${name}-${theme}.webp`)
}

/** Takes a still in both themes; prepare puts the app in the state to show. */
async function stills(page, name, prepare, clip) {
  for (const theme of ['light', 'dark']) {
    await setTheme(page, theme)
    await prepare(page)
    await still(page, name, theme, typeof clip === 'function' ? await clip(page) : clip)
  }
}

/** Records act as docs/img/<name>-<theme>.webp, an animation, in both themes. */
async function animation(page, name, prepare, act) {
  for (const theme of ['light', 'dark']) {
    await setTheme(page, theme)
    await prepare(page)
    mouse = { x: WIDTH / 2, y: HEIGHT / 2 }
    await page.mouse.move(mouse.x, mouse.y)
    const webm = path.join(work, `${name}-${theme}.webm`)
    const rec = await page.screencast({ path: webm, scale: 1 })
    await sleep(600)
    await act(page)
    await sleep(1200)
    await rec.stop()
    execFileSync('ffmpeg', ['-y', '-loglevel', 'error', '-i', webm, '-vf', `fps=15,scale=${WIDTH}:-1:flags=lanczos`,
      '-loop', '0', '-c:v', 'libwebp_anim', '-quality', '75', '-compression_level', '6',
      path.join(out, `${name}-${theme}.webp`)])
    console.log(`  ${name}-${theme}.webp (animated)`)
  }
}

// ── Scenes ────────────────────────────────────────────────────────────────────────────

const scenes = {}

scenes.start = (page) =>
  stills(page, 'start-page', async (p) => {
    await reset(p, [])
  })

scenes.repo = async (page) => {
  await stills(page, 'repository', async (p) => {
    await reset(p)
    await p.waitForSelector(byText('src/alerts.js', 'button'))
    await sleep(500)
  })
  // The tour's copy, with numbered markers on each part of the window. Positions are in the
  // still's pixels (2560 x 1600).
  const marks = [
    [400, 45], // tabs
    [250, 1100], // sidebar
    [1480, 124], // toolbar
    [1180, 205], // working tree
    [1450, 780], // graph
    [2490, 130], // right panel
    [2470, 1075], // commit message
    [2030, 40], // command palette
  ]
  for (const theme of ['light', 'dark']) annotate('repository', theme, marks)
}

/** Writes <name>-annotated-<theme>.webp: the still with numbered rose markers at marks. */
function annotate(name, theme, marks) {
  const draw = marks.flatMap(([x, y], i) => [
    '-fill', '#cb0162', '-stroke', '#ffffff', '-strokewidth', '4', '-draw', `circle ${x},${y} ${x + 27},${y}`,
    '-fill', '#ffffff', '-stroke', 'none', '-pointsize', '32', '-annotate', `+${x - (i >= 9 ? 18 : 9)}+${y + 11}`, String(i + 1),
  ])
  run('magick', [path.join(out, `${name}-${theme}.webp`), '-font', 'Arial-Bold', ...draw, '-quality', '90',
    path.join(out, `${name}-annotated-${theme}.webp`)])
  console.log(`  ${name}-annotated-${theme}.webp`)
}

scenes.details = (page) =>
  stills(page, 'commit-details', async (p) => {
    await reset(p, ['tidewater'], { clean: true })
    await selectCommit(p, 'Predict tide height')
  })

scenes.palette = (page) =>
  stills(page, 'command-palette', async (p) => {
    await reset(p, ['tidewater'], { clean: true })
    await p.keyboard.down('Control')
    await p.keyboard.press('k')
    await p.keyboard.up('Control')
    await sleep(300)
    await p.keyboard.type('pull')
    await sleep(400)
  })

scenes.diff = (page) =>
  stills(page, 'diff-unified', async (p) => {
    await reset(p)
    await click(p, byText('src/alerts.js', 'button'))
    await sleep(800)
  })

/** Opens file in the commit with this subject, in the File layout, rendered. */
async function preview(p, subject, file) {
  await reset(p, ['tidewater'], { clean: true })
  await selectCommit(p, subject)
  await click(p, byText(file, 'button'))
  await sleep(700)
  await click(p, byText('File', 'button'))
  const rendered = await p.$(byText('Rendered', 'button'))
  if (rendered) await rendered.click()
  await sleep(1500)
}

scenes.csv = (page) => stills(page, 'preview-csv', (p) => preview(p, 'Add the station list as CSV', 'data/stations.csv'))
scenes.mermaid = (page) => stills(page, 'preview-mermaid', (p) => preview(p, 'Diagram how an alert travels', 'docs/alerts.mmd'))
scenes.markdown = (page) => stills(page, 'preview-markdown', (p) => preview(p, 'Document the units', 'README.md'))
scenes.json = (page) => stills(page, 'preview-json', (p) => preview(p, 'Add tidal constants', 'data/constants.json'))

scenes.blame = (page) =>
  stills(page, 'blame', async (p) => {
    await reset(p, ['tidewater'], { clean: true })
    await selectCommit(p, 'Show wind in knots')
    await click(p, byText('src/feed.js', 'button'))
    await sleep(600)
    await click(p, byText('Blame', 'button'))
    await sleep(1200)
  })

scenes.dashboard = (page) =>
  stills(page, 'group-dashboard', async (p) => {
    await reset(p, [])
    await click(p, byText('Dashboard'))
    await sleep(1500)
  })

scenes.settings = (page) =>
  stills(page, 'settings', async (p) => {
    await reset(p, ['tidewater'], { clean: true })
    await p.keyboard.down('Control')
    await p.keyboard.press(',')
    await p.keyboard.up('Control')
    await sleep(800)
  })

scenes.reflog = (page) =>
  stills(page, 'reflog', async (p) => {
    await reset(p)
    await p.keyboard.down('Control')
    await p.keyboard.press('k')
    await p.keyboard.up('Control')
    await sleep(300)
    await p.keyboard.type('reflog')
    await p.keyboard.press('Enter')
    await sleep(1000)
  })

/** Right-clicks row i of the graph and chooses the menu item containing text. */
async function rowMenu(p, i, text) {
  const r = await row(p, i)
  await clickAt(p, r.x, r.y, { button: 'right', pause: 500 })
  await click(p, byText(text), { pause: 700 })
}

/** Opens the interactive rebase editor for feature/offline-cache, from where it left main.
 * Row 7 is main's tip: above it are origin/main, the five branch commits and the stash. */
async function openRebase(p) {
  await rowMenu(p, 7, 'Interactive rebase')
  await p.waitForSelector(byText('Start rebase', 'button'))
  await sleep(500)
}

scenes.rebase = (page) =>
  stills(page, 'rebase-editor', async (p) => {
    await reset(p, ['tidewater'], { clean: true, branch: 'feature/offline-cache' })
    await openRebase(p)
  })

/** Sets the action for the commit with this subject in the rebase editor. */
async function action(p, subject, value) {
  const sel = `select[aria-label="Action for ${subject}"]`
  await click(p, sel, { pause: 200 })
  await p.keyboard.press('Escape')
  await p.select(sel, value)
  await sleep(500)
}

scenes.rebaseAnim = (page) =>
  animation(
    page,
    'rebase',
    (p) => reset(p, ['tidewater'], { clean: true, branch: 'feature/offline-cache' }),
    async (p) => {
      await openRebase(p)
      await action(p, 'debugging', 'drop')
      await action(p, 'fixup! Cache the last reading from each station', 'fixup')
      await action(p, 'wip', 'squash')
      await sleep(600)
      await click(p, byText('Start rebase', 'button'), { pause: 4000 })
    },
  )

/** Merges feature/metric-units into main, which stops on a conflict in src/feed.js. */
async function merge(p) {
  await click(p, `::-p-xpath(//button[contains(@class, 'ref') and normalize-space(.)='feature/metric-units'])`, {
    button: 'right',
    pause: 500,
  })
  await click(p, byText('Merge feature/metric-units into main'), { pause: 300 })
  // The merge stops on the conflict: wait for its banner, and put away git's own message.
  await p.waitForSelector('.op')
  await sleep(1000)
  if (await p.$(byText('Dismiss', 'button'))) await click(p, byText('Dismiss', 'button'), { pause: 600 })
}

scenes.conflict = (page) =>
  stills(page, 'conflict-tool', async (p) => {
    await reset(p, ['tidewater'], { clean: true })
    await merge(p)
    await click(p, 'button.wip', { pause: 800 })
    await click(p, '[title="Open the conflict tool"]', { pause: 1500 })
  })

scenes.conflictAnim = (page) =>
  animation(
    page,
    'conflict',
    (p) => reset(p, ['tidewater'], { clean: true }),
    async (p) => {
      await merge(p)
      await sleep(800)
      await click(p, 'button.wip', { pause: 900 })
      await click(p, '[title="Open the conflict tool"]', { pause: 1600 })
      await click(p, byText('Use theirs', 'button'), { pause: 1200 })
      await click(p, byText('Mark resolved', 'button'), { pause: 1500 })
      const go = await p.$(byText('Continue', 'button'))
      if (go) await click(p, byText('Continue', 'button'), { pause: 2000 })
    },
  )

scenes.commitAnim = (page) =>
  animation(
    page,
    'stage-and-commit',
    (p) => reset(p),
    async (p) => {
      await click(p, byText('src/alerts.js', 'button'), { pause: 1200 })
      await click(p, byText('Stage hunk', 'button'), { pause: 1000 })
      await keys(p, 'Esc', { hold: 700 })
      await click(p, 'textarea[aria-label="Commit summary"], input[aria-label="Commit summary"]', { pause: 300 })
      await type(p, 'Soften the wind alert and add a knots helper')
      await sleep(500)
      await keys(p, 'Ctrl+Enter', { hold: 1800 })
    },
  )

scenes.undoAnim = (page) =>
  animation(
    page,
    'undo',
    (p) => reset(p, ['tidewater'], { clean: true }),
    async (p) => {
      await click(p, `::-p-xpath(//button[contains(@class, 'ref') and normalize-space(.)='feature/offline-cache'])`, {
        button: 'right',
        pause: 500,
      })
      await click(p, byText('Merge feature/offline-cache into main'), { pause: 300 })
      // Give the merge time to land; the toolbar's label can lag behind it (see #220).
      await sleep(2500)
      await keys(p, 'Ctrl+Z', { hold: 5000 })
    },
  )

scenes.paletteAnim = (page) =>
  animation(
    page,
    'palette',
    (p) => reset(p, ['tidewater'], { clean: true }),
    async (p) => {
      await keys(p, 'Ctrl+K', { hold: 500 })
      await type(p, 'pull reb', 90)
      await sleep(900)
      await keys(p, 'Esc', { hold: 400 })
      await keys(p, 'Ctrl+K', { hold: 400 })
      await type(p, 'fetch', 90)
      await sleep(500)
      await keys(p, 'Enter', { hold: 1800 })
    },
  )

scenes.searchAnim = (page) =>
  animation(
    page,
    'search',
    (p) => reset(p, ['tidewater'], { clean: true }),
    async (p) => {
      await keys(p, 'Ctrl+F', { hold: 400 })
      await type(p, 'Closes: #14', 80)
      await sleep(400)
      await keys(p, 'Enter', { hold: 1500 })
      await keys(p, 'Ctrl+F', { hold: 300 })
      await p.keyboard.down('Control')
      await p.keyboard.press('a')
      await p.keyboard.up('Control')
      await type(p, 'Priya', 80)
      await keys(p, 'Enter', { hold: 900 })
      await keys(p, 'Enter', { hold: 900 })
      await keys(p, 'Enter', { hold: 1200 })
    },
  )

// ── Main ──────────────────────────────────────────────────────────────────────────────

mkdirSync(out, { recursive: true })
rmSync(work, { recursive: true, force: true })
mkdirSync(work, { recursive: true })

const browser = await puppeteer.launch({
  executablePath: CHROME,
  headless: true,
  defaultViewport: { width: WIDTH, height: HEIGHT, deviceScaleFactor: 2 },
  args: ['--hide-scrollbars', '--font-render-hinting=none'],
})
try {
  const page = await browser.newPage()
  await page.goto(APP, { waitUntil: 'networkidle0' })
  const wanted = process.argv.slice(2)
  for (const [name, scene] of Object.entries(scenes)) {
    if (wanted.length && !wanted.includes(name)) continue
    console.log(name)
    try {
      await scene(page)
    } catch (e) {
      await page.screenshot({ path: path.join(work, `failed-${name}.png`) })
      console.error(`  failed; the page is in ${path.join(work, `failed-${name}.png`)}`)
      throw e
    }
  }
} finally {
  await browser.close()
}
