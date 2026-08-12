'use client'

import { useEffect, useState } from 'react'
import { type Checkpoint, faroURL, fetchCheckpoint } from '@/lib/checkpoint'

type State =
  | { status: 'loading' }
  | { status: 'ready'; checkpoint: Checkpoint }
  | { status: 'error'; message: string }

export default function Home() {
  const [state, setState] = useState<State>({ status: 'loading' })

  useEffect(() => {
    // `cancelled` guards against a response arriving after the component has
    // gone away, which would otherwise set state on an unmounted tree.
    let cancelled = false
    fetchCheckpoint()
      .then((checkpoint) => {
        if (!cancelled) setState({ status: 'ready', checkpoint })
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'unknown error',
          })
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <>
      <h1>Checkpoint actual</h1>
      <p className="lede">
        Un checkpoint es la afirmación firmada del log sobre su propio estado: cuántas
        entradas contiene y cuál es la raíz de su árbol de Merkle.
      </p>

      {state.status === 'loading' && <p>Consultando {faroURL}…</p>}

      {state.status === 'error' && (
        <div className="notice">
          <strong>No se pudo obtener el checkpoint</strong>
          {state.message}
          <p>
            Verificá que el log esté corriendo en <code>{faroURL}</code> y que{' '}
            <code>FARO_CORS_ORIGIN</code> permita el origen de esta página.
          </p>
        </div>
      )}

      {state.status === 'ready' && (
        <>
          <dl className="field-grid">
            <dt>Origen</dt>
            <dd>{state.checkpoint.origin}</dd>
            <dt>Entradas</dt>
            <dd>{state.checkpoint.size.toString()}</dd>
            <dt>Raíz de Merkle</dt>
            <dd>{state.checkpoint.rootHash}</dd>
            <dt>Firmas</dt>
            <dd>{state.checkpoint.signatures.length}</dd>
          </dl>

          <div className="notice">
            <strong>Estos datos todavía no están verificados</strong>
            La página muestra el checkpoint tal como lo devolvió el servidor. La
            verificación de la firma y de las pruebas de inclusión y consistencia se ejecuta
            en el navegador y es el próximo paso de este explorador. Hasta que exista, no
            hay ninguna afirmación de validez acá: mostrar un sello de «verificado» sin
            haber verificado sería peor que no mostrar nada.
          </div>

          <h2>Nota firmada</h2>
          <pre>{state.checkpoint.raw}</pre>
        </>
      )}

      <h2>Verificar por tu cuenta</h2>
      <p>
        No hace falta confiar en esta página. El log expone la API de lectura{' '}
        <a href="https://c2sp.org/tlog-tiles">tlog-tiles</a> como archivos estáticos, y la
        CLI <code>faro-verify</code> comprueba lo mismo desde tu propia máquina.
      </p>
    </>
  )
}
