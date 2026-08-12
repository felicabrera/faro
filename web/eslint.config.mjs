// Flat ESLint config.
//
// Next.js 16 removed the `next lint` subcommand, so package.json runs `eslint .`
// directly against this file. A `"lint": "next lint"` script would fail, and it
// is a required status check, so leave it as it is.
import next from 'eslint-config-next'

const config = [
  {
    ignores: ['.next/**', 'out/**', 'node_modules/**', 'next-env.d.ts'],
  },
  ...next,
]

export default config
