// API base URL is read from VITE_API_BASE in .env files.
// .env.development  → VITE_API_BASE=/api   (proxied by vite dev server)
// .env.test         → VITE_API_BASE=https://test-api.example.com
// .env.production   → VITE_API_BASE=https://api.example.com

export const API_BASE = import.meta.env.VITE_API_BASE as string ?? ''
