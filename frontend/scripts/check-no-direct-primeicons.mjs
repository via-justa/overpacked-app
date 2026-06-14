import fs from 'node:fs/promises'
import path from 'node:path'

const workspaceRoot = path.resolve(process.cwd())
const srcRoot = path.join(workspaceRoot, 'src')

const ignoredDirs = new Set(['node_modules', 'dist', '.git', '.vscode'])

// Patterns to detect direct PrimeIcons usage
const patterns = [
  {
    regex: /\bicon=["']pi\s+pi-[a-z-]+["']/g,
    description: 'Button icon prop with direct PrimeIcons (use iconRegistry instead)',
  },
  {
    regex: /\bclass=["']pi\s+pi-[a-z-]+/g,
    description: 'Direct PrimeIcons class (use AppIcon component instead)',
  },
  {
    regex: /\bicon:\s*["']pi\s+pi-[a-z-]+["']/g,
    description: 'Icon in data object with direct PrimeIcons (use iconRegistry template literal)',
  },
  {
    regex: /:class=["']pi\s+pi-[a-z-]+/g,
    description: 'Dynamic class binding with direct PrimeIcons (use iconRegistry or AppIcon)',
  },
]

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

    if (entry.isFile() && (fullPath.endsWith('.vue') || fullPath.endsWith('.ts') || fullPath.endsWith('.js'))) {
      discoveredFiles.push(fullPath)
    }
  }

  return discoveredFiles
}

const findViolations = async (filesToScan) => {
  const violations = []

  for (const filePath of filesToScan) {
    const content = await fs.readFile(filePath, 'utf8')

    for (const pattern of patterns) {
      for (const match of content.matchAll(pattern.regex)) {
        const index = match.index ?? 0
        const line = content.slice(0, index).split('\n').length

        violations.push({ 
          filePath, 
          line, 
          token: match[0], 
          description: pattern.description 
        })
      }
    }
  }

  return violations
}

const main = async () => {
  const filesToScan = await listScannableFiles(srcRoot)
  const violations = await findViolations(filesToScan)

  if (violations.length === 0) {
    console.log('✅ No direct PrimeIcons usage found. All icons use iconRegistry or AppIcon component.')
    return
  }

  console.error('❌ Direct PrimeIcons usage detected. Use iconRegistry or AppIcon component instead:\n')

  for (const violation of violations) {
    const relativePath = path.relative(workspaceRoot, violation.filePath)
    console.error(`  ${relativePath}:${violation.line}`)
    console.error(`    ${violation.token}`)
    console.error(`    ${violation.description}\n`)
  }

  console.error('\nMigration guide:')
  console.error('  - Button icons: :icon="`pi ${iconRegistry.category.name}`"')
  console.error('  - Inline icons: <AppIcon category="category" name="name" />')
  console.error('  - Data arrays: icon: `pi ${iconRegistry.category.name}`')
  console.error('\nSee /frontend/src/lib/icons/README.md for details.')

  process.exitCode = 1
}

await main()
