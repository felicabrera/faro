# Política de Seguridad — FARO

FARO es el registro público de transparencia y auditoría del sistema de voto
electrónico ÁGORA: un log append-only basado en árboles de Merkle (construido
sobre Tessera). La integridad verificable de este log es la base sobre la que
descansa la auditabilidad de todo el sistema, por lo que tratamos cualquier
falla que la comprometa como crítica.

## Reportar una vulnerabilidad

Si encontrás una vulnerabilidad de seguridad, una debilidad criptográfica o un
problema de integridad, reportalo **de forma privada** usando la función de
reporte privado de GitHub:

> Pestaña **Security** → **Report a vulnerability**

No abras un issue público ni lo divulgues en redes antes de que exista un
arreglo disponible.

- **Acuse de recibo:** dentro de las 72 horas.
- **Plan de resolución:** dentro de los 7 días desde el acuse de recibo.
- **Divulgación coordinada:** se espera que nos permitas publicar el arreglo
  antes de cualquier divulgación pública. Si el reporte es válido, damos crédito
  público al reportante salvo que prefiera permanecer anónimo.

## Alcance

**Dentro de alcance**

- Integridad del log de transparencia: garantías de append-only, imposibilidad
  de modificación retroactiva, correctitud de la construcción del árbol de
  Merkle.
- Verificación de pruebas de inclusión y de consistencia.
- Firma y verificación de checkpoints / signed tree heads.
- Autenticación, autorización y manejo de tokens de la API de FARO.
- Configuración de infraestructura y de CI/CD que pueda afectar la integridad
  de los artefactos publicados.

**Fuera de alcance**

- Vulnerabilidades en dependencias de terceros sin un vector de explotación
  demostrable en este proyecto (reportalas upstream).
- Denegación de servicio por volumen de tráfico.
- Hallazgos automatizados de escáneres sin análisis de impacto.
- Ingeniería social contra el equipo o la institución.

## Versiones soportadas

| Versión        | Soportada |
| -------------- | --------- |
| `main` (última)| ✅        |

Este es un proyecto académico en desarrollo activo (Proyecto Final de Grado,
Ingeniería en Informática, Uruguay). No hay versiones estables anteriores con
soporte.

---

# Security Policy (English)

FARO is the public transparency and audit log of the ÁGORA electronic voting
system: an append-only, Merkle-tree-based log built on Tessera. The verifiable
integrity of this log underpins the auditability of the entire system, so any
flaw that compromises it is treated as critical.

## Reporting a Vulnerability

If you discover a security vulnerability, cryptographic weakness, or integrity
issue, please report it **privately** through GitHub's private vulnerability
reporting feature (**Security** tab → **Report a vulnerability**) rather than
opening a public issue.

- **Acknowledgement:** within 72 hours.
- **Resolution timeline:** within 7 days of acknowledgement.
- **Coordinated disclosure** is expected; please allow us to ship a fix before
  public disclosure. Valid reports are credited publicly unless you prefer to
  remain anonymous.

## Scope

**In scope:** append-only guarantees and resistance to retroactive
modification, Merkle tree construction, inclusion and consistency proof
verification, checkpoint / signed-tree-head signing, FARO API authentication
and token handling, and infrastructure or CI/CD configuration affecting the
integrity of published artifacts.

**Out of scope:** third-party dependency issues with no demonstrated exploit
path in this project, volumetric denial of service, unanalyzed automated
scanner output, and social engineering.

## Supported Versions

| Version         | Supported |
| --------------- | --------- |
| `main` (latest) | ✅        |
