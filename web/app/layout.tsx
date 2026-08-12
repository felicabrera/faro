import type { Metadata } from 'next'
import Link from 'next/link'
import type { ReactNode } from 'react'
import './globals.css'

export const metadata: Metadata = {
  title: 'FARO — Registro público auditable',
  description:
    'Explorador público del log append-only de ÁGORA: checkpoints firmados y pruebas de inclusión y consistencia verificables por cualquiera.',
}

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="es">
      <body>
        <header className="site-header">
          <Link className="wordmark" href="/">
            FARO
          </Link>
          <span className="tagline">Registro público auditable</span>
        </header>
        <main>{children}</main>
        <footer className="site-footer">
          <p>
            Proyecto Final de Grado, Ingeniería en Informática, Universidad Católica del
            Uruguay. Código bajo AGPL-3.0.
          </p>
        </footer>
      </body>
    </html>
  )
}
