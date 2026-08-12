# FARO

**Formally Auditable Registry Operations** — el registro público, append-only y verificable
del sistema de votación [ÁGORA](https://github.com/felicabrera/agora).

> **English (short version).** FARO is the public transparency log that makes ÁGORA's
> results independently auditable. It is a [Tessera](https://github.com/transparency-dev/tessera)
> append-only Merkle log served over the [C2SP tlog-tiles](https://c2sp.org/tlog-tiles) read
> API, plus a Go CLI and a web explorer for checking it. Anyone can verify that a ballot is
> in the log and that the log was never rewritten, without trusting the operator. This
> repository currently contains the log service, the CLI skeleton and the explorer shell;
> the verification subcommands are under development. Documentation below is in Spanish.

---

## Qué es

FARO responde una sola pregunta, y la responde de forma que no haga falta creernos:
**¿este voto cifrado está en el registro, y el registro fue alterado?**

- **Prueba de inclusión**: demuestra que una entrada está en un árbol con raíz `R`, con
  ~log₂(n) hashes. Es lo que permite a un votante confirmar "mi voto quedó registrado".
- **Prueba de consistencia**: demuestra que un árbol anterior es prefijo de uno posterior,
  sin reordenar, editar ni borrar. Es lo que permite a un auditor confirmar "el log no fue
  manipulado".

El costo de auditar crece con `log n`, no con `n`: verificar una inclusión en un log de un
millón de entradas cuesta el mismo puñado de hashes que en uno de mil.

FARO **no sabe nada de papeletas**. Almacena bytes opacos y prueba que no los alteró. Que
esos bytes sean un voto válido es una conclusión de ÁGORA. Esa frontera es deliberada: hace
que FARO pueda auditarse por separado.

## Estado actual

| Componente | Estado |
|---|---|
| `cmd/faro-log` | Servicio operativo: `POST /add`, API tlog-tiles, checkpoints firmados |
| `internal/log` | Configuración del appender Tessera sobre el driver POSIX |
| `cmd/faro-verify` | CLI: `version`, `keygen` |
| `web/` | Explorador: muestra el checkpoint actual |

Pendiente, en orden de trabajo: subcomandos de verificación (`checkpoint`, `inclusion`,
`consistency`, `monitor`), verificación en el navegador, testigos independientes
(C2SP tlog-witness) y autenticación del camino de escritura.

## Por qué Tessera y no un Merkle propio

FARO no implementa un árbol de Merkle. Configura Tessera, la biblioteca de
transparency-dev que también sostiene logs de Certificate Transparency en producción, y
sirve el resultado en el formato tlog-tiles.

Escribir el árbol nosotros significaría pedirle a un auditor que confíe en una
implementación a medida del único componente del que depende todo lo demás. Usando la misma
biblioteca y el mismo formato de cable que CT, un auditor puede apuntar verificadores ya
existentes y escritos por terceros contra FARO.

La API de lectura son **archivos estáticos**. Cada byte que un auditor necesita está en
disco en un formato documentado y estándar, así que el log puede espejarse, archivarse o
servirse desde cualquier servidor web sin ejecutar este binario. Un log de transparencia
cuyo contenido solo puede leerse a través de su propia API es uno en el que hay que confiar;
este no lo es.

## Quickstart

Requiere **Go 1.26+** y **Node 20.9+**.

```console
$ git clone https://github.com/felicabrera/faro && cd faro
$ make build

# 1. Generar la identidad del log. El nombre es el origin que los verificadores fijan.
$ ./bin/faro-verify keygen --name faro.local/dev

# 2. Levantar el log con la clave privada que imprimió el paso anterior.
$ export FARO_SIGNING_KEY='PRIVATE+KEY+faro.local/dev+...'
$ export FARO_CORS_ORIGIN=http://localhost:3000
$ ./bin/faro-log

# 3. En otra terminal: agregar entradas y leer el checkpoint.
$ curl -X POST --data-binary 'hola' localhost:2025/add
0
$ curl localhost:2025/checkpoint
faro.local/dev
1
7wRDNzWQ3aSAKuUCUQ0MDbFOhFPYxLp5nJUXHTEy0Cg=

— faro.local/dev ...
```

Y el explorador:

```console
$ make web-install
$ make web-dev        # http://localhost:3000
```

`make help` lista el resto de los targets.

## Configuración

El servicio se configura por entorno. La clave de firma no tiene default y no se genera al
arrancar: un log que inventara una clave al no encontrarla arrancaría sano y produciría
checkpoints que nadie puede verificar contra la clave publicada, que es una falla peor que
no arrancar.

| Variable | Default | Qué hace |
|---|---|---|
| `FARO_SIGNING_KEY` | — (requerida) | Clave privada de checkpoints |
| `FARO_STORAGE_DIR` | `./data/log` | Raíz del log en disco |
| `FARO_ADDR` | `:2025` | Dirección de escucha |
| `FARO_CHECKPOINT_INTERVAL` | `10s` | Cada cuánto se publica una raíz |
| `FARO_BATCH_MAX_AGE` | `1s` | Cuánto espera una entrada a secuenciarse |
| `FARO_BATCH_MAX_SIZE` | `256` | Cuántas entradas se secuencian juntas |
| `FARO_CORS_ORIGIN` | — | Único origen que puede leer el log desde un navegador |
| `FARO_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

## Estructura del repositorio

```
cmd/faro-log/        # servicio del log: POST /add + API tlog-tiles
cmd/faro-verify/     # CLI pública de auditoría
internal/log/        # configuración del appender Tessera (driver POSIX)
internal/config/     # configuración por entorno
internal/version/    # información de build estampada por el toolchain
web/                 # explorador público (Next.js, exportación estática)
```

## Límites conocidos

- **El camino de escritura todavía no está autenticado.** Hoy cualquiera que alcance el
  puerto puede agregar entradas. La autenticación llega con la integración de ÁGORA.
- **No hay testigos independientes todavía.** El protocolo C2SP tlog-witness convierte la
  auditoría multipartita en una propiedad del protocolo en lugar de un acuerdo
  institucional; está previsto y no implementado.
- **El explorador todavía no verifica.** Muestra el checkpoint tal como lo devolvió el
  servidor y lo dice explícitamente en la página. La verificación en el navegador es el
  próximo paso.

## Seguridad

Ver [`SECURITY.md`](SECURITY.md). Para reportar una vulnerabilidad, usar el canal de
divulgación coordinada descrito allí y no un issue público.

## Licencia

GNU AGPL-3.0 — ver [`LICENSE`](LICENSE).

## Referencias

[RFC 6962 (Certificate Transparency)](https://www.rfc-editor.org/rfc/rfc6962) ·
[C2SP tlog-tiles](https://c2sp.org/tlog-tiles) ·
[C2SP tlog-witness](https://c2sp.org/tlog-witness) ·
[Tessera](https://github.com/transparency-dev/tessera) ·
[signed notes](https://pkg.go.dev/golang.org/x/mod/sumdb/note)
