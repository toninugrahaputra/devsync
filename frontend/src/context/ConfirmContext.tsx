import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { AlertTriangle } from 'lucide-react'

interface ConfirmOptions {
  title?: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

type ConfirmFn = (options: ConfirmOptions | string) => Promise<boolean>

const ConfirmContext = createContext<ConfirmFn | null>(null)

interface PendingConfirm {
  options: ConfirmOptions
  resolve: (value: boolean) => void
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirm | null>(null)

  const confirm = useCallback<ConfirmFn>((options) => {
    const normalized = typeof options === 'string' ? { message: options } : options
    return new Promise<boolean>((resolve) => {
      setPending({ options: normalized, resolve })
    })
  }, [])

  function settle(result: boolean) {
    pending?.resolve(result)
    setPending(null)
  }

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      {pending &&
        createPortal(
          <div
            className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100] p-4"
            onClick={() => settle(false)}
          >
            <div
              className="bg-white rounded-lg shadow-xl w-full max-w-sm p-5"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="flex items-start gap-3">
                {pending.options.danger && (
                  <span className="shrink-0 mt-0.5 text-red-500">
                    <AlertTriangle size={20} />
                  </span>
                )}
                <div>
                  {pending.options.title && (
                    <h2 className="text-sm font-semibold text-gray-900 mb-1">{pending.options.title}</h2>
                  )}
                  <p className="text-sm text-gray-600">{pending.options.message}</p>
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-5">
                <button
                  onClick={() => settle(false)}
                  className="text-sm px-3 py-1.5 rounded text-gray-600 hover:bg-gray-100"
                >
                  {pending.options.cancelLabel || 'Cancel'}
                </button>
                <button
                  onClick={() => settle(true)}
                  autoFocus
                  className={`text-sm px-3 py-1.5 rounded font-medium ${
                    pending.options.danger
                      ? 'bg-red-600 text-white hover:bg-red-700'
                      : 'bg-[#D6AE32] text-gray-900 hover:bg-[#B6942B]'
                  }`}
                >
                  {pending.options.confirmLabel || 'Confirm'}
                </button>
              </div>
            </div>
          </div>,
          document.body
        )}
    </ConfirmContext.Provider>
  )
}

export function useConfirm() {
  const ctx = useContext(ConfirmContext)
  if (!ctx) throw new Error('useConfirm must be used within ConfirmProvider')
  return ctx
}
