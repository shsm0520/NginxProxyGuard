import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  getSystemSettings,
  updateSystemSettings,
  testACME,
} from '../api/settings';
import type { SystemSettings, UpdateSystemSettingsRequest } from '../types/settings';
import { HelpTip } from './common/HelpTip';

export default function SSLACMESettings() {
  const { t } = useTranslation('settings');
  const queryClient = useQueryClient();
  const [editedSettings, setEditedSettings] = useState<UpdateSystemSettingsRequest>({});
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const { data: settings, isLoading } = useQuery({
    queryKey: ['systemSettings'],
    queryFn: getSystemSettings,
  });

  const updateMutation = useMutation({
    mutationFn: updateSystemSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systemSettings'] });
      setEditedSettings({});
      setSaveMessage({ type: 'success', text: t('system.acme.saveSuccess') });
      setTimeout(() => setSaveMessage(null), 3000);
    },
    onError: (error: Error) => {
      setSaveMessage({ type: 'error', text: t('system.acme.saveFailed', { error: error.message }) });
      setTimeout(() => setSaveMessage(null), 5000);
    },
  });

  const testACMEMutation = useMutation({
    mutationFn: testACME,
  });

  const handleChange = (key: keyof UpdateSystemSettingsRequest, value: string | number | boolean | object) => {
    setEditedSettings((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = () => {
    if (Object.keys(editedSettings).length > 0) {
      updateMutation.mutate(editedSettings);
    }
  };

  const getValue = <K extends keyof SystemSettings>(key: K): SystemSettings[K] | undefined => {
    if (key in editedSettings) {
      return (editedSettings as Partial<SystemSettings>)[key] as SystemSettings[K];
    }
    return settings?.[key];
  };

  const isModified = Object.keys(editedSettings).length > 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const inputClass = "mt-1 w-full px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-xl font-bold text-slate-800 dark:text-white">{t('system.tabs.acme')}</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            {t('system.acme.enable.description')}
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={!isModified || updateMutation.isPending}
          className="px-4 py-2 text-[13px] font-semibold bg-blue-600 text-white hover:bg-blue-700 rounded-lg disabled:opacity-50 disabled:bg-slate-300 transition-colors"
        >
          {updateMutation.isPending ? t('system.buttons.saving') : t('system.buttons.save')}
        </button>
      </div>

      {/* Save Message */}
      {saveMessage && (
        <div className={`px-4 py-3 rounded-lg text-sm font-medium ${saveMessage.type === 'success'
          ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800'
          : 'bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
          }`}>
          {saveMessage.text}
        </div>
      )}

      {/* Test ACME Result */}
      {testACMEMutation.data && (
        <div className={`rounded-xl shadow-sm border p-4 flex items-start gap-3 transition-colors ${testACMEMutation.data.status === 'ok' ? 'border-emerald-200 bg-emerald-50 dark:bg-emerald-900/10 dark:border-emerald-800' :
          testACMEMutation.data.status === 'warning' ? 'border-amber-200 bg-amber-50 dark:bg-amber-900/10 dark:border-amber-800' :
            'border-slate-200 bg-slate-50 dark:bg-slate-700/50 dark:border-slate-600'
          }`}>
          <div className={`p-1 rounded-full ${testACMEMutation.data.status === 'ok' ? 'bg-emerald-100 dark:bg-emerald-900/30' :
            testACMEMutation.data.status === 'warning' ? 'bg-amber-100 dark:bg-amber-900/30' : 'bg-slate-100 dark:bg-slate-700'
            }`}>
            {testACMEMutation.data.status === 'ok' && (
              <svg className="w-5 h-5 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            )}
            {testACMEMutation.data.status === 'warning' && (
              <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            )}
            {testACMEMutation.data.status === 'disabled' && (
              <svg className="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
          </div>
          <p className={`text-sm font-medium ${testACMEMutation.data.status === 'ok' ? 'text-emerald-700 dark:text-emerald-400' :
            testACMEMutation.data.status === 'warning' ? 'text-amber-700 dark:text-amber-400' : 'text-slate-600 dark:text-slate-400'
            }`}>
            {t(`system.acme.testResult.${testACMEMutation.data.status}`, { defaultValue: testACMEMutation.data.message })}
          </p>
        </div>
      )}

      {/* Configuration */}
      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5 space-y-6 transition-colors">
        {/* Enable/Disable */}
        <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-200 dark:border-slate-600 transition-colors">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={getValue('acme_enabled') ?? true}
              onChange={(e) => handleChange('acme_enabled', e.target.checked)}
              className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 dark:bg-slate-700"
            />
            <div>
              <span className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.acme.enable.label')}</span>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {t('system.acme.enable.description')}
              </p>
            </div>
          </label>
        </div>

        {/* ACME Email */}
        <div>
          <label className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
            {t('system.acme.email.label')}
            <HelpTip contentKey="help.acme.email" ns="settings" />
          </label>
          <input
            type="email"
            value={(editedSettings.acme_email ?? settings?.acme_email) || ''}
            onChange={(e) => handleChange('acme_email', e.target.value)}
            className={inputClass}
            placeholder={t('system.acme.email.placeholder')}
          />
          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
            {t('system.acme.email.help')}
          </p>
        </div>

        {/* Staging Mode */}
        <div className="py-3 px-4 bg-amber-50 dark:bg-amber-900/10 rounded-lg border border-amber-200 dark:border-amber-800 transition-colors">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              checked={getValue('acme_staging') ?? false}
              onChange={(e) => handleChange('acme_staging', e.target.checked)}
              className="mt-0.5 w-5 h-5 rounded border-amber-300 dark:border-amber-700 text-amber-600 focus:ring-amber-500 focus:ring-offset-0 dark:bg-slate-700"
            />
            <div>
              <span className="text-sm font-semibold text-amber-800 dark:text-amber-400 flex items-center gap-2">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                </svg>
                {t('system.acme.staging.label')}
                <HelpTip contentKey="help.acme.staging" ns="settings" />
              </span>
              <p className="text-xs text-amber-700 dark:text-amber-500 mt-1">
                {t('system.acme.staging.description')}
              </p>
            </div>
          </label>
        </div>

        {/* Auto Renew */}
        <div className="space-y-5 border-t border-slate-200 dark:border-slate-700 pt-6">
          <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
            <svg className="w-4 h-4 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            {t('system.acme.autoRenew.title')}
          </h3>

          <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/50 rounded-lg border border-slate-200 dark:border-slate-600 transition-colors">
            <label className="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={getValue('acme_auto_renew') ?? true}
                onChange={(e) => handleChange('acme_auto_renew', e.target.checked)}
                className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 dark:bg-slate-700"
              />
              <div>
                <span className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.acme.autoRenew.enableLabel')}</span>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  {t('system.acme.autoRenew.enableDescription')}
                </p>
              </div>
            </label>
          </div>

          <div>
            <label className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
              {t('system.acme.autoRenew.daysBeforeLabel')}
              <HelpTip contentKey="help.acme.renewBefore" ns="settings" />
            </label>
            <div className="flex items-center gap-2 mt-1">
              <input
                type="number"
                min="1"
                max="60"
                value={getValue('acme_renew_days_before') ?? 30}
                onChange={(e) => handleChange('acme_renew_days_before', parseInt(e.target.value))}
                className="w-24 px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors text-right"
              />
              <span className="text-sm text-slate-500 dark:text-slate-400">{t('system.acme.autoRenew.daysBeforeHelp')}</span>
            </div>
            <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
              {t('system.acme.autoRenew.help')}
            </p>
          </div>
        </div>

        {/* Test Button */}
        <div className="pt-6 border-t border-slate-200 dark:border-slate-700">
          <button
            onClick={() => testACMEMutation.mutate()}
            disabled={testACMEMutation.isPending}
            className="px-4 py-2.5 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 font-medium rounded-lg text-sm transition-colors flex items-center gap-2 disabled:opacity-50"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {testACMEMutation.isPending ? t('system.buttons.test.testing') : t('system.acme.testButton')}
          </button>
        </div>
      </div>
    </div>
  );
}
