import { useEffect, useRef } from 'react'

// Minimal shape of the Google Identity Services global — loaded via the
// <script> tag in index.html, not an npm package, so it's not otherwise typed.
declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: { client_id: string; callback: (resp: { credential: string }) => void }) => void
          renderButton: (parent: HTMLElement, options: Record<string, unknown>) => void
        }
      }
    }
  }
}

const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID as string | undefined

interface Props {
  onCredential: (credential: string) => void
}

export default function GoogleSignInButton({ onCredential }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!GOOGLE_CLIENT_ID || !containerRef.current) return

    let cancelled = false
    // The GIS script loads async; poll briefly until window.google is ready.
    function render() {
      if (cancelled) return
      if (!window.google || !containerRef.current) {
        setTimeout(render, 100)
        return
      }
      window.google.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID!,
        callback: (resp) => onCredential(resp.credential),
      })
      window.google.accounts.id.renderButton(containerRef.current, {
        theme: 'outline',
        size: 'large',
        width: 320,
        text: 'continue_with',
      })
    }
    render()

    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (!GOOGLE_CLIENT_ID) return null

  return <div ref={containerRef} className="flex justify-center" />
}
