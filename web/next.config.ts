import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // The explorer is a static site. It reads the log over HTTP from the browser
  // and holds no server-side state, which is what lets it be served from a CDN
  // and, more importantly, what makes it impossible for the site to lie about a
  // proof without the visitor's own browser being complicit.
  output: 'export',
  reactStrictMode: true,
  // The log lives on a different origin, so its URL is build-time configuration.
  env: {
    NEXT_PUBLIC_FARO_URL: process.env.NEXT_PUBLIC_FARO_URL ?? 'http://localhost:2025',
  },
}

export default nextConfig
