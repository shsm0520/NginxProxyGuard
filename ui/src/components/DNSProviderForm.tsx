import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { createDNSProvider, updateDNSProvider, testDNSProvider } from '../api/dns-providers'
import { useTranslation } from 'react-i18next'
import { HelpTip } from './common/HelpTip'
import type { DNSProvider, CreateDNSProviderRequest } from '../types/certificate'

interface DNSProviderFormProps {
  provider?: DNSProvider | null
  onClose: () => void
  onSuccess: () => void
}

type ProviderType = 'cloudflare' | 'route53' | 'manual'

export default function DNSProviderForm({ provider, onClose, onSuccess }: DNSProviderFormProps) {
  const { t } = useTranslation('certificates')
  const isEditing = !!provider

  const [name, setName] = useState(provider?.name || '')
  const [providerType, setProviderType] = useState<ProviderType>(
    (provider?.provider_type as ProviderType) || 'cloudflare'
  )
  const [isDefault, setIsDefault] = useState(provider?.is_default || false)
  const [error, setError] = useState('')
  const [testResult, setTestResult] = useState<'success' | 'failed' | null>(null)
  const [isTesting, setIsTesting] = useState(false)

  // Cloudflare credentials
  const [cfApiToken, setCfApiToken] = useState('')
  const [cfApiKey, setCfApiKey] = useState('')
  const [cfEmail, setCfEmail] = useState('')
  const [cfZoneId, setCfZoneId] = useState('')

  // Route53 credentials
  const [awsAccessKeyId, setAwsAccessKeyId] = useState('')
  const [awsSecretAccessKey, setAwsSecretAccessKey] = useState('')
  const [awsRegion, setAwsRegion] = useState('us-east-1')
  const [awsHostedZoneId, setAwsHostedZoneId] = useState('')

  const buildCredentials = (): Record<string, string> => {
    if (providerType === 'cloudflare') {
      const creds: Record<string, string> = {}
      if (cfApiToken) creds.api_token = cfApiToken
      if (cfApiKey) creds.api_key = cfApiKey
      if (cfEmail) creds.email = cfEmail
      if (cfZoneId) creds.zone_id = cfZoneId
      return creds
    }
    if (providerType === 'route53') {
      const creds: Record<string, string> = {}
      if (awsAccessKeyId) creds.access_key_id = awsAccessKeyId
      if (awsSecretAccessKey) creds.secret_access_key = awsSecretAccessKey
      if (awsRegion) creds.region = awsRegion
      if (awsHostedZoneId) creds.hosted_zone_id = awsHostedZoneId
      return creds
    }
    return {}
  }

  const createMutation = useMutation({
    mutationFn: createDNSProvider,
    onSuccess: () => {
      onSuccess()
    },
    onError: (err: Error) => {
      setError(err.message || t('dnsProviders.form.errors.createFailed'))
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateDNSProviderRequest> }) =>
      updateDNSProvider(id, data),
    onSuccess: () => {
      onSuccess()
    },
    onError: (err: Error) => {
      setError(err.message || t('dnsProviders.form.errors.updateFailed'))
    },
  })

  const handleTest = async () => {
    setTestResult(null)
    setError('')
    setIsTesting(true)

    const data: CreateDNSProviderRequest = {
      name,
      provider_type: providerType,
      credentials: buildCredentials(),
      is_default: isDefault,
    }

    try {
      const success = await testDNSProvider(data)
      setTestResult(success ? 'success' : 'failed')
    } catch {
      setTestResult('failed')
    } finally {
      setIsTesting(false)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name.trim()) {
      setError(t('dnsProviders.form.errors.nameRequired'))
      return
    }

    const data: CreateDNSProviderRequest = {
      name: name.trim(),
      provider_type: providerType,
      credentials: buildCredentials(),
      is_default: isDefault,
    }

    if (isEditing && provider) {
      updateMutation.mutate({ id: provider.id, data })
    } else {
      createMutation.mutate(data)
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto border dark:border-slate-700">
        <div className="p-6 border-b border-slate-200 dark:border-slate-700">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-slate-900 dark:text-white">
              {isEditing ? t('dnsProviders.form.editTitle') : t('dnsProviders.form.addTitle')}
            </h2>
            <button
              onClick={onClose}
              disabled={isPending}
              className="text-slate-400 hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {error && (
            <div className="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-4 text-red-700 dark:text-red-400 text-sm">
              {error}
            </div>
          )}

          {testResult === 'success' && (
            <div className="bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 rounded-lg p-4 text-green-700 dark:text-green-400 text-sm flex items-center gap-2">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              {t('dnsProviders.form.test.success')}
            </div>
          )}

          {testResult === 'failed' && (
            <div className="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-4 text-red-700 dark:text-red-400 text-sm flex items-center gap-2">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              {t('dnsProviders.form.test.failed')}
            </div>
          )}

          {/* Provider Name */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              {t('dnsProviders.form.providerName')}
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('dnsProviders.form.providerNamePlaceholder')}
              className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 placeholder:text-slate-400 bg-white dark:bg-slate-700 dark:text-white"
              required
            />
          </div>

          {/* Provider Type */}
          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">
              {t('dnsProviders.form.providerType')}
            </label>
            <div className="grid grid-cols-3 gap-3">
              <button
                type="button"
                onClick={() => setProviderType('cloudflare')}
                className={`relative flex flex-col items-center p-4 rounded-xl border-2 transition-all ${providerType === 'cloudflare'
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 bg-white dark:bg-slate-800'
                  }`}
              >
                <div className={`w-10 h-10 rounded-full flex items-center justify-center mb-2 ${providerType === 'cloudflare' ? 'bg-orange-100 dark:bg-orange-900/30' : 'bg-slate-100 dark:bg-slate-700'
                  }`}>
                  <svg className={`w-5 h-5 ${providerType === 'cloudflare' ? 'text-orange-600 dark:text-orange-400' : 'text-slate-500 dark:text-slate-400'}`} fill="currentColor" viewBox="0 0 24 24">
                    <path d="M16.5 11.38l-.21.2-.14.27-.42.16-.3.09c-.2.16-.37.36-.52.58-.2.3-.32.63-.35.98h-.02l-.17.33-.2.23c-.1.08-.22.14-.34.18-.13.03-.27.04-.4.04-.08 0-.16-.02-.24-.05a.4.4 0 01-.16-.12h-.02l-.17-.33-.13-.14c-.04-.08-.06-.17-.06-.26-.01-.08-.01-.16.01-.24.03-.08.09-.14.16-.17.08-.06.17-.1.27-.12.14-.04.3-.06.46-.06h.01c.06 0 .1-.03.14-.08a.2.2 0 00.05-.14v-.02a.2.2 0 00-.05-.14.2.2 0 00-.14-.08h-.01c-.22 0-.43.03-.63.1-.2.06-.38.15-.54.27-.15.12-.28.27-.38.44-.1.17-.17.36-.21.57-.04.21-.05.42-.02.63.03.21.1.41.2.58.1.17.23.32.39.43.16.12.34.2.54.25.2.05.41.07.62.05.2-.02.4-.07.58-.16.18-.08.34-.2.47-.34.14-.14.24-.3.32-.48.08-.18.13-.37.15-.57.02-.2.01-.4-.04-.59a1.2 1.2 0 00-.23-.51c-.1-.16-.24-.3-.4-.4-.15-.1-.32-.18-.5-.23a2 2 0 00-.56-.08h-.01a.2.2 0 01-.14-.05.2.2 0 01-.05-.14v-.05c0-.05.02-.1.05-.13a.2.2 0 01.14-.06h.01c.22-.02.43-.08.62-.18.2-.1.36-.24.5-.4.14-.17.24-.36.3-.57.07-.21.1-.43.09-.65a1.7 1.7 0 00-.13-.63 1.5 1.5 0 00-.34-.52 1.5 1.5 0 00-.52-.34c-.2-.08-.41-.13-.63-.13-.22 0-.43.05-.63.13-.2.08-.37.2-.52.34-.15.15-.26.32-.34.52-.08.2-.13.41-.13.63 0 .22.05.44.13.64.08.2.2.37.34.52.15.14.32.26.52.34.2.08.41.13.63.15h.02c.06 0 .1.02.13.06.04.03.05.08.05.13v.03a.2.2 0 01-.05.14.2.2 0 01-.13.05h-.02c-.3.02-.58.11-.84.26-.25.15-.47.35-.64.59-.17.24-.3.51-.38.8a2.4 2.4 0 00.02 1.24c.09.3.23.57.42.8.19.24.42.44.68.59.27.15.56.25.87.3.31.05.63.05.94 0 .31-.05.6-.15.87-.3.26-.15.5-.35.68-.59.19-.23.33-.5.42-.8a2.4 2.4 0 00.02-1.24c-.08-.29-.2-.56-.38-.8a2.1 2.1 0 00-.64-.59c-.26-.15-.55-.24-.84-.26" />
                  </svg>
                </div>
                <span className={`text-sm font-medium ${providerType === 'cloudflare' ? 'text-primary-700 dark:text-primary-300' : 'text-slate-700 dark:text-slate-300'}`}>
                  Cloudflare
                </span>
              </button>

              <button
                type="button"
                onClick={() => setProviderType('route53')}
                className={`relative flex flex-col items-center p-4 rounded-xl border-2 transition-all ${providerType === 'route53'
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 bg-white dark:bg-slate-800'
                  }`}
              >
                <div className={`w-10 h-10 rounded-full flex items-center justify-center mb-2 ${providerType === 'route53' ? 'bg-amber-100 dark:bg-amber-900/30' : 'bg-slate-100 dark:bg-slate-700'
                  }`}>
                  <svg className={`w-5 h-5 ${providerType === 'route53' ? 'text-amber-600 dark:text-amber-400' : 'text-slate-500 dark:text-slate-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
                  </svg>
                </div>
                <span className={`text-sm font-medium ${providerType === 'route53' ? 'text-primary-700 dark:text-primary-300' : 'text-slate-700 dark:text-slate-300'}`}>
                  AWS Route 53
                </span>
              </button>

              <button
                type="button"
                onClick={() => setProviderType('manual')}
                className={`relative flex flex-col items-center p-4 rounded-xl border-2 transition-all ${providerType === 'manual'
                  ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 bg-white dark:bg-slate-800'
                  }`}
              >
                <div className={`w-10 h-10 rounded-full flex items-center justify-center mb-2 ${providerType === 'manual' ? 'bg-slate-200 dark:bg-slate-600' : 'bg-slate-100 dark:bg-slate-700'
                  }`}>
                  <svg className={`w-5 h-5 ${providerType === 'manual' ? 'text-slate-700 dark:text-slate-300' : 'text-slate-500 dark:text-slate-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </div>
                <span className={`text-sm font-medium ${providerType === 'manual' ? 'text-primary-700 dark:text-primary-300' : 'text-slate-700 dark:text-slate-300'}`}>
                  {t('dnsProviders.types.manual')}
                </span>
              </button>
            </div>
          </div>

          {/* Cloudflare Credentials */}
          {providerType === 'cloudflare' && (
            <div className="space-y-4 bg-slate-50 dark:bg-slate-700/50 rounded-xl p-5 border border-slate-200 dark:border-slate-600">
              <h3 className="text-sm font-medium text-slate-700 dark:text-slate-200 flex items-center gap-2">
                <svg className="w-5 h-5 text-orange-500" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M16.5 11.38l-.21.2-.14.27-.42.16-.3.09c-.2.16-.37.36-.52.58-.2.3-.32.63-.35.98h-.02l-.17.33-.2.23c-.1.08-.22.14-.34.18-.13.03-.27.04-.4.04-.08 0-.16-.02-.24-.05a.4.4 0 01-.16-.12" />
                </svg>
                {t('dnsProviders.form.cloudflare.title')}
              </h3>

              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 flex items-center gap-2">
                  {t('dnsProviders.form.cloudflare.apiToken')}
                  <HelpTip content={t('dnsProviders.form.cloudflare.apiTokenHelp')} />
                </label>
                <input
                  type="password"
                  value={cfApiToken}
                  onChange={(e) => setCfApiToken(e.target.value)}
                  placeholder={t('dnsProviders.form.cloudflare.apiTokenPlaceholder')}
                  className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 placeholder:text-slate-400 bg-white dark:bg-slate-700 dark:text-white"
                />
                <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                  {t('dnsProviders.form.cloudflare.apiTokenHelp')}
                </p>
              </div>

              <div className="border-t border-slate-200 dark:border-slate-700 pt-4">
                <p className="text-xs text-slate-500 dark:text-slate-400 mb-3">{t('dnsProviders.form.cloudflare.orGlobalKey')}</p>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                      {t('dnsProviders.form.cloudflare.apiKey')}
                    </label>
                    <input
                      type="password"
                      value={cfApiKey}
                      onChange={(e) => setCfApiKey(e.target.value)}
                      className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                      {t('dnsProviders.form.cloudflare.email')}
                    </label>
                    <input
                      type="email"
                      value={cfEmail}
                      onChange={(e) => setCfEmail(e.target.value)}
                      className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 flex items-center gap-2">
                  {t('dnsProviders.form.cloudflare.zoneId')}
                  <HelpTip content={t('dnsProviders.form.cloudflare.zoneIdPlaceholder')} />
                </label>
                <input
                  type="text"
                  value={cfZoneId}
                  onChange={(e) => setCfZoneId(e.target.value)}
                  placeholder={t('dnsProviders.form.cloudflare.zoneIdPlaceholder')}
                  className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 placeholder:text-slate-400 bg-white dark:bg-slate-700 dark:text-white"
                />
              </div>
            </div>
          )}

          {/* Route53 Credentials */}
          {providerType === 'route53' && (
            <div className="space-y-4 bg-slate-50 dark:bg-slate-700/50 rounded-xl p-5 border border-slate-200 dark:border-slate-600">
              <h3 className="text-sm font-medium text-slate-700 dark:text-slate-200 flex items-center gap-2">
                <svg className="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
                </svg>
                {t('dnsProviders.form.route53.title')}
              </h3>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {t('dnsProviders.form.route53.accessKeyId')}
                  </label>
                  <input
                    type="text"
                    value={awsAccessKeyId}
                    onChange={(e) => setAwsAccessKeyId(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {t('dnsProviders.form.route53.secretAccessKey')}
                  </label>
                  <input
                    type="password"
                    value={awsSecretAccessKey}
                    onChange={(e) => setAwsSecretAccessKey(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {t('dnsProviders.form.route53.region')}
                  </label>
                  <input
                    type="text"
                    value={awsRegion}
                    onChange={(e) => setAwsRegion(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {t('dnsProviders.form.route53.hostedZoneId')}
                  </label>
                  <input
                    type="text"
                    value={awsHostedZoneId}
                    onChange={(e) => setAwsHostedZoneId(e.target.value)}
                    className="w-full rounded-lg border border-slate-300 dark:border-slate-600 px-4 py-2.5 text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-slate-700 dark:text-white"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Manual DNS info */}
          {providerType === 'manual' && (
            <div className="bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-xl p-5">
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div>
                  <h4 className="text-sm font-medium text-amber-800 dark:text-amber-300">{t('dnsProviders.form.manual.title')}</h4>
                  <p className="mt-1 text-sm text-amber-700 dark:text-amber-400">
                    {t('dnsProviders.form.manual.desc')}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Is Default Toggle */}
          <div className="flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-700/50 rounded-xl border border-transparent dark:border-slate-700">
            <div>
              <h4 className="text-sm font-medium text-slate-700 dark:text-slate-300">{t('dnsProviders.form.setDefault')}</h4>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{t('dnsProviders.form.setDefaultDesc')}</p>
            </div>
            <button
              type="button"
              onClick={() => setIsDefault(!isDefault)}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:focus:ring-offset-slate-800 ${isDefault ? 'bg-primary-600' : 'bg-slate-200 dark:bg-slate-600'
                }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${isDefault ? 'translate-x-5' : 'translate-x-0'
                  }`}
              />
            </button>
          </div>

          {/* Actions */}
          <div className="flex justify-between pt-4 border-t border-slate-200 dark:border-slate-700">
            <button
              type="button"
              onClick={handleTest}
              disabled={isPending || isTesting || providerType === 'manual'}
              className="px-4 py-2 text-sm font-medium text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800 rounded-lg hover:bg-primary-100 dark:hover:bg-primary-900/40 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 flex items-center gap-2"
            >
              {isTesting && (
                <svg className="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
              )}

              {t('dnsProviders.form.test.button')}
            </button>

            <div className="flex gap-3">
              <button
                type="button"
                onClick={onClose}
                disabled={isPending}
                className="px-4 py-2 text-sm font-medium text-slate-700 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white"
              >
                {t('dnsProviders.form.cancel')}
              </button>
              <button
                type="submit"
                disabled={isPending}
                className="bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
              >
                {isPending && (
                  <svg className="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                )}
                {isEditing ? t('dnsProviders.form.update') : t('dnsProviders.form.create')}
              </button>
            </div>
          </div>
        </form>
      </div >
    </div >
  )
}
