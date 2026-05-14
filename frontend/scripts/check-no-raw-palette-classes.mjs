import fs from 'node:fs/promises'
import path from 'node:path'

const workspaceRoot = path.resolve(process.cwd())
const srcRoot = path.join(workspaceRoot, 'src')
const styleFilePath = path.join(srcRoot, 'style.css')

const ignoredDirs = new Set(['node_modules', 'dist', '.git', '.vscode'])
const allowedFiles = new Set([styleFilePath])

const palettePattern = /\b(?:bg|text|border|ring|shadow|decoration|divide|from|via|to|file:bg|file:text|hover:file:bg)-(?:stone|emerald|rose|red|amber|yellow|orange|sky|white)(?:-[0-9]+(?:\/[0-9]+)?)?\b/g // NOSONAR

const listScannableFiles = async (dir) => {
  const discoveredFiles = []
  const entries = await fs.readdir(dir, { withFileTypes: true })

  for (const entry of entries) {
    if (ignoredDirs.has(entry.name)) {
      continue
    }

    const fullPath = path.join(dir, entry.name)

    if (entry.isDirectory()) {
      const childFiles = await listScannableFiles(fullPath)
      discoveredFiles.push(...childFiles)
      continue
    }

    if (entry.isFile() && (fullPath.endsWith('.vue') || fullPath.endsWith('.css'))) {
      discoveredFiles.push(fullPath)
    }
  }

  return discoveredFiles
}

const findViolations = async (filesToScan) => {
  const violations = []

  for (const filePath of filesToScan) {
    if (allowedFiles.has(filePath)) {
      continue
    }

    const content = await fs.readFile(filePath, 'utf8')

    for (const match of content.matchAll(palettePattern)) {
      const index = match.index ?? 0
      const line = content.slice(0, index).split('\n').length

      violations.push({ filePath, line, token: match[0] })
    }
  }

  return violations
}

const main = async () => {
  const filesToScan = await listScannableFiles(srcRoot)
  const violations = await findViolations(filesToScan)

  if (violations.length === 0) {
    return
  }

  console.error('Raw Tailwind palette utilities found. Use semantic tokens instead:')

  for (const violation of violations) {
    console.error(`- ${path.relative(workspaceRoot, violation.filePath)}:${violation.line} -> ${violation.token}`)
  }

  process.exitCode = 1
}

await main()