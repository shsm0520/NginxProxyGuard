import type { CreateProxyHostRequest } from '../../../types/proxy-host'
import { useTranslation } from 'react-i18next'
import { HelpTip } from '../../common/HelpTip'

interface PerformanceTabProps {
  formData: CreateProxyHostRequest
  setFormData: React.Dispatch<React.SetStateAction<CreateProxyHostRequest>>
}

export function PerformanceTabContent({ formData, setFormData }: PerformanceTabProps) {
  const { t } = useTranslation('proxyHost')
  return (
    <div className="space-y-6">
      {/* Caching */}
      <div className={`p-4 rounded-lg border-2 transition-colors ${formData.cache_enabled ? 'bg-indigo-50 border-indigo-200 dark:bg-indigo-900/20 dark:border-indigo-800' : 'bg-slate-50 border-slate-200 dark:bg-slate-800/50 dark:border-slate-700'
        }`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center ${formData.cache_enabled ? 'bg-indigo-100 dark:bg-indigo-900/40' : 'bg-slate-200 dark:bg-slate-700'
              }`}>
              <svg className={`w-5 h-5 ${formData.cache_enabled ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-400 dark:text-slate-500'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
              </svg>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex items-center h-5 mt-1">
                <input
                  type="checkbox"
                  checked={formData.cache_enabled}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, cache_enabled: e.target.checked }))
                  }
                  className="w-4 h-4 text-indigo-600 rounded border-slate-300 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
                  {t('form.performance.cache.enabled')}
                  <HelpTip contentKey="help.performance.cache" />
                </label>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  {t('form.performance.cache.description')}
                </p>
              </div>
            </div>
          </div>
        </div>
        {formData.cache_enabled && (
          <div className="mt-3 text-xs text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 p-3 rounded border border-slate-200 dark:border-slate-700">
            <p className="font-medium mb-1">Cache settings:</p>
            <ul className="list-disc list-inside space-y-0.5 text-slate-500 dark:text-slate-400">
              <li>200, 301, 302 responses cached for 1 hour</li>
              <li>404 responses cached for 1 minute</li>
              <li>X-Cache-Status header shows HIT/MISS</li>
            </ul>
          </div>
        )}
      </div>

      {/* WebSocket */}
      <div className={`p-4 rounded-lg border-2 transition-colors ${formData.allow_websocket_upgrade ? 'bg-cyan-50 border-cyan-200 dark:bg-cyan-900/20 dark:border-cyan-800' : 'bg-slate-50 border-slate-200 dark:bg-slate-800/50 dark:border-slate-700'
        }`}>
        <label className="flex items-center justify-between cursor-pointer">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center ${formData.allow_websocket_upgrade ? 'bg-cyan-100 dark:bg-cyan-900/40' : 'bg-slate-200 dark:bg-slate-700'
              }`}>
              <svg className={`w-5 h-5 ${formData.allow_websocket_upgrade ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400 dark:text-slate-500'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <div>
              <span className="text-sm font-medium text-slate-900 dark:text-white flex items-center gap-2">
                {t('form.basic.websocket')}
                <HelpTip contentKey="help.performance.websocket" />
              </span>
              <p className="text-xs text-slate-500 dark:text-slate-400">{t('form.basic.websocketDescription')}</p>
            </div>
          </div>
          <input
            type="checkbox"
            checked={formData.allow_websocket_upgrade}
            onChange={(e) =>
              setFormData((prev) => ({
                ...prev,
                allow_websocket_upgrade: e.target.checked,
              }))
            }
            className="rounded border-slate-300 text-cyan-600 focus:ring-cyan-500 h-5 w-5"
          />
        </label>
      </div>
    </div>
  )
}
