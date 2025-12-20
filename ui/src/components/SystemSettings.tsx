import { useState } from 'react';
import { useTranslation, Trans } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getSystemSettings,
  updateSystemSettings,
  getGeoIPStatus,
  triggerGeoIPUpdate,
  testACME,
  getLogFiles,
  viewLogFile,
  downloadLogFile,
  deleteLogFile,
  triggerLogRotation,
  getSystemLogConfig,
  updateSystemLogConfig,
} from '../api/settings';
import type { SystemSettings as SystemSettingsType, UpdateSystemSettingsRequest, LogFileInfo, SystemLogConfig } from '../types/settings';
import { HelpTip } from './common/HelpTip';
import { formatBytes } from './log-viewer/utils';

type TabType = 'geoip' | 'acme' | 'botfilter' | 'security' | 'maintenance' | 'logfiles' | 'systemlogs';

export default function SystemSettings() {
  const queryClient = useQueryClient();
  const { t, i18n } = useTranslation('settings');
  const [activeTab, setActiveTab] = useState<TabType>('geoip');
  const [editedSettings, setEditedSettings] = useState<UpdateSystemSettingsRequest>({});
  const [editedLogConfig, setEditedLogConfig] = useState<Partial<SystemLogConfig>>({});
  const [excludePatternsText, setExcludePatternsText] = useState<string | null>(null);
  const [showLicenseKey, setShowLicenseKey] = useState(false);
  const [viewingFile, setViewingFile] = useState<string | null>(null);
  const [viewContent, setViewContent] = useState<string>('');
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [_downloadError, setDownloadError] = useState<string | null>(null);

  const { data: settings, isLoading } = useQuery({
    queryKey: ['systemSettings'],
    queryFn: getSystemSettings,
  });

  const { data: geoipStatus, refetch: refetchGeoIP } = useQuery({
    queryKey: ['geoipStatus'],
    queryFn: getGeoIPStatus,
  });

  const updateMutation = useMutation({
    mutationFn: updateSystemSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systemSettings'] });
      queryClient.invalidateQueries({ queryKey: ['geoipStatus'] });
      setEditedSettings({});
    },
  });

  const geoipUpdateMutation = useMutation({
    mutationFn: () => triggerGeoIPUpdate(true),
    onSuccess: () => {
      setTimeout(() => refetchGeoIP(), 5000);
    },
  });

  const testACMEMutation = useMutation({
    mutationFn: testACME,
  });

  // Log files management
  const { data: logFilesData, refetch: refetchLogFiles } = useQuery({
    queryKey: ['logFiles'],
    queryFn: getLogFiles,
    enabled: activeTab === 'logfiles',
  });

  const viewFileMutation = useMutation({
    mutationFn: ({ filename, lines }: { filename: string; lines: number }) =>
      viewLogFile(filename, lines),
    onSuccess: (data) => {
      setViewContent(data.content);
    },
  });

  const deleteFileMutation = useMutation({
    mutationFn: deleteLogFile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['logFiles'] });
      setConfirmDelete(null);
    },
  });

  const rotateMutation = useMutation({
    mutationFn: triggerLogRotation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['logFiles'] });
    },
  });


  // System Log Config
  const { data: systemLogConfig } = useQuery({
    queryKey: ['systemLogConfig'],
    queryFn: getSystemLogConfig,
    enabled: activeTab === 'systemlogs',
  });

  const updateLogConfigMutation = useMutation({
    mutationFn: updateSystemLogConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systemLogConfig'] });
      setEditedLogConfig({});
    },
  });
  const handleChange = (key: keyof UpdateSystemSettingsRequest, value: string | number | boolean | object) => {
    setEditedSettings((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = () => {
    if (Object.keys(editedSettings).length > 0) {
      updateMutation.mutate(editedSettings);
    }
    if (Object.keys(editedLogConfig).length > 0 && systemLogConfig) {
      updateLogConfigMutation.mutate({
        ...systemLogConfig,
        ...editedLogConfig,
        levels: {
          ...systemLogConfig.levels,
          ...(editedLogConfig.levels || {}),
        }
      });
    }
  };

  const getValue = <K extends keyof SystemSettingsType>(key: K): SystemSettingsType[K] | undefined => {
    if (key in editedSettings) {
      return (editedSettings as Partial<SystemSettingsType>)[key] as SystemSettingsType[K];
    }
    return settings?.[key];
  };

  const getLogConfigValue = <K extends keyof SystemLogConfig>(key: K): SystemLogConfig[K] | undefined => {
    if (key === 'levels') {
      return undefined; // Handle levels separately
    }
    if (key in editedLogConfig) {
      return (editedLogConfig as Partial<SystemLogConfig>)[key] as SystemLogConfig[K];
    }
    return systemLogConfig?.[key];
  };

  const isModified = Object.keys(editedSettings).length > 0 || Object.keys(editedLogConfig).length > 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const tabs: { id: TabType; label: string; icon: React.ReactNode }[] = [
    { id: 'geoip', label: t('system.tabs.geoip'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg> },
    { id: 'acme', label: t('system.tabs.acme'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg> },
    { id: 'botfilter', label: t('system.tabs.botfilter'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg> },
    { id: 'security', label: t('system.tabs.security'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg> },
    { id: 'maintenance', label: t('system.tabs.maintenance'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg> },
    { id: 'logfiles', label: t('system.tabs.logfiles'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg> },
    { id: 'systemlogs', label: t('system.tabs.systemlogs', 'System Logs'), icon: <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg> },
  ];

  const handleViewFile = (filename: string) => {
    setViewingFile(filename);
    setViewContent('');
    viewFileMutation.mutate({ filename, lines: 200 });
  };

  const handleDownloadFile = async (filename: string) => {
    try {
      await downloadLogFile(filename);
    } catch {
      setDownloadError(t('rawLogs.downloadFailed') || 'Download failed');
      setTimeout(() => setDownloadError(null), 5000);
    }
  };

  // Common input class
  const inputClass = "mt-1 w-full px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-xl font-bold text-slate-800 dark:text-white">{t('system.title')}</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            {t('system.geoip.enable.description')}
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={!isModified || updateMutation.isPending}
          className="px-4 py-2 text-[13px] font-semibold bg-blue-600 text-white hover:bg-blue-700 rounded-lg disabled:opacity-50 disabled:bg-slate-300 transition-colors"
        >
          {updateMutation.isPending || updateLogConfigMutation.isPending ? t('system.buttons.saving') : t('system.buttons.save')}
        </button>
      </div>

      {/* Tabs */}
      <div className="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700">
        <div className="border-b border-slate-200 dark:border-slate-700 px-2">
          <div className="flex overflow-x-auto gap-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-4 py-3 text-[13px] font-semibold whitespace-nowrap border-b-2 transition-all ${activeTab === tab.id
                  ? 'border-blue-600 text-blue-600 dark:text-blue-400 dark:border-blue-400'
                  : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700'
                  }`}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        <div className="p-6">
          {/* GeoIP Settings */}
          {activeTab === 'geoip' && (
            <div className="space-y-6">
              {/* Status Card */}
              <div className={`p-5 rounded-xl ${geoipStatus?.status === 'active' ? 'bg-emerald-50 dark:bg-emerald-900/10 border border-emerald-200 dark:border-emerald-900/30' :
                geoipStatus?.status === 'error' ? 'bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-900/30' :
                  'bg-slate-50 dark:bg-slate-700/30 border border-slate-200 dark:border-slate-700'
                }`}>
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-white">{t('system.geoip.status.title')}</h3>
                    <p className="text-sm mt-1.5">
                      {geoipStatus?.status === 'active' && (
                        <span className="text-emerald-700 dark:text-emerald-400">
                          {t('system.geoip.status.active')}
                          {geoipStatus.last_updated && t('system.geoip.status.lastUpdated', { date: new Date(geoipStatus.last_updated).toLocaleDateString() })}
                        </span>
                      )}
                      {geoipStatus?.status === 'inactive' && (
                        <span className="text-slate-600 dark:text-slate-400">{t('system.geoip.status.inactive')}</span>
                      )}
                      {geoipStatus?.status === 'error' && (
                        <span className="text-red-700 dark:text-red-400">{geoipStatus.error_message || t('system.geoip.status.error')}</span>
                      )}
                      {geoipStatus?.status === 'updating' && (
                        <span className="text-blue-700 dark:text-blue-400">{t('system.geoip.status.updating')}</span>
                      )}
                    </p>
                  </div>
                  <button
                    onClick={() => geoipUpdateMutation.mutate()}
                    disabled={geoipUpdateMutation.isPending || !settings?.maxmind_license_key}
                    className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:bg-slate-300 transition-colors"
                  >
                    {geoipUpdateMutation.isPending ? t('system.buttons.updating') : t('system.buttons.updateNow')}
                  </button>
                </div>
                {geoipStatus && (
                  <div className="mt-4 flex gap-4 text-sm">
                    <span className={`flex items-center gap-1.5 ${geoipStatus.country_db ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400 dark:text-slate-500'}`}>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        {geoipStatus.country_db ? <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /> : <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />}
                      </svg>
                      {t('system.geoip.status.countryDb')}
                    </span>
                    <span className={`flex items-center gap-1.5 ${geoipStatus.asn_db ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400 dark:text-slate-500'}`}>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        {geoipStatus.asn_db ? <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /> : <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />}
                      </svg>
                      {t('system.geoip.status.asnDb')}
                    </span>
                  </div>
                )}
              </div>

              {/* Enable/Disable */}
              <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                <label className="flex items-start gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={getValue('geoip_enabled') ?? false}
                    onChange={(e) => handleChange('geoip_enabled', e.target.checked)}
                    className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                  />
                  <div>
                    <span className="text-sm font-semibold text-slate-700 dark:text-slate-200 flex items-center gap-2">
                      {t('system.geoip.enable.label')}
                      <HelpTip contentKey="help.geoip.enable" ns="settings" />
                    </span>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                      {t('system.geoip.enable.description')}
                    </p>
                  </div>
                </label>
              </div>

              {/* MaxMind Credentials */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
                    {t('system.geoip.account.idLabel')}
                    <HelpTip contentKey="help.geoip.account" ns="settings" />
                  </label>
                  <input
                    type="text"
                    value={(editedSettings.maxmind_account_id ?? settings?.maxmind_account_id) || ''}
                    onChange={(e) => handleChange('maxmind_account_id', e.target.value)}
                    className={inputClass}
                    placeholder={t('system.geoip.account.idPlaceholder')}
                  />
                  <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                    <a href="https://www.maxmind.com/en/geolite2/signup" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:underline font-medium">
                      {t('system.geoip.account.createAccount')}
                    </a>
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.geoip.account.keyLabel')}</label>
                  <div className="relative mt-1">
                    <input
                      type={showLicenseKey ? 'text' : 'password'}
                      value={(editedSettings.maxmind_license_key ?? settings?.maxmind_license_key) || ''}
                      onChange={(e) => handleChange('maxmind_license_key', e.target.value)}
                      className="w-full px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors pr-20"
                      placeholder={t('system.geoip.account.keyPlaceholder')}
                    />
                    <button
                      type="button"
                      onClick={() => setShowLicenseKey(!showLicenseKey)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 px-2 py-1 text-xs font-medium text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 bg-slate-100 dark:bg-slate-600 hover:bg-slate-200 dark:hover:bg-slate-500 rounded transition-colors"
                    >
                      {showLicenseKey ? t('system.buttons.hide') : t('system.buttons.view')}
                    </button>
                  </div>
                  <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                    {t('system.geoip.account.keyHelp')}
                  </p>
                </div>
              </div>

              {/* Auto Update */}
              <div className="space-y-5 border-t border-slate-200 dark:border-slate-700 pt-6">
                <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
                  <svg className="w-4 h-4 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  {t('system.geoip.autoUpdate.title')}
                </h3>

                <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={getValue('geoip_auto_update') ?? true}
                      onChange={(e) => handleChange('geoip_auto_update', e.target.checked)}
                      className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                    />
                    <div>
                      <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">{t('system.geoip.autoUpdate.enableLabel')}</span>
                      <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {t('system.geoip.autoUpdate.enableDescription')}
                      </p>
                    </div>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.geoip.autoUpdate.intervalLabel')}</label>
                  <select
                    value={(editedSettings.geoip_update_interval ?? settings?.geoip_update_interval) || '7d'}
                    onChange={(e) => handleChange('geoip_update_interval', e.target.value)}
                    className={`${inputClass} md:w-48`}
                  >
                    <option value="1d">{t('system.geoip.autoUpdate.intervals.daily')}</option>
                    <option value="7d">{t('system.geoip.autoUpdate.intervals.weekly')}</option>
                    <option value="30d">{t('system.geoip.autoUpdate.intervals.monthly')}</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {/* ACME Settings */}
          {activeTab === 'acme' && (
            <div className="space-y-6">
              {/* Test ACME Result */}
              {testACMEMutation.data && (
                <div className={`p-4 rounded-xl flex items-start gap-3 ${testACMEMutation.data.status === 'ok' ? 'bg-emerald-50 dark:bg-emerald-900/10 border border-emerald-200 dark:border-emerald-900/30' :
                  testACMEMutation.data.status === 'warning' ? 'bg-amber-50 dark:bg-amber-900/10 border border-amber-200 dark:border-amber-900/30' :
                    'bg-slate-50 dark:bg-slate-700/30 border border-slate-200 dark:border-slate-700'
                  }`}>
                  <div className={`p-1 rounded-full ${testACMEMutation.data.status === 'ok' ? 'bg-emerald-100 dark:bg-emerald-900/30' :
                    testACMEMutation.data.status === 'warning' ? 'bg-amber-100 dark:bg-amber-900/30' : 'bg-slate-100 dark:bg-slate-700/50'
                    }`}>
                    {testACMEMutation.data.status === 'ok' && (
                      <svg className="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    )}
                    {testACMEMutation.data.status === 'warning' && (
                      <svg className="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                    )}
                    {testACMEMutation.data.status === 'disabled' && (
                      <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    )}
                  </div>
                  <p className={`text-sm font-medium ${testACMEMutation.data.status === 'ok' ? 'text-emerald-700 dark:text-emerald-400' :
                    testACMEMutation.data.status === 'warning' ? 'text-amber-700 dark:text-amber-400' : 'text-slate-600 dark:text-slate-400'
                    }`}>
                    {testACMEMutation.data.message}
                  </p>
                </div>
              )}

              {/* Enable/Disable */}
              <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                <label className="flex items-start gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={getValue('acme_enabled') ?? true}
                    onChange={(e) => handleChange('acme_enabled', e.target.checked)}
                    className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                  />
                  <div>
                    <span className="text-sm font-semibold text-slate-700 dark:text-slate-200 flex items-center gap-2">
                      {t('system.acme.enable.label')}
                      <HelpTip contentKey="help.acme.enable" ns="settings" />
                    </span>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                      {t('system.acme.enable.description')}
                    </p>
                  </div>
                </label>
              </div>

              {/* ACME Email */}
              <div>
                <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-2">
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
              <div className="py-3 px-4 bg-amber-50 dark:bg-amber-900/10 rounded-lg border border-amber-200 dark:border-amber-900/30">
                <label className="flex items-start gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={getValue('acme_staging') ?? false}
                    onChange={(e) => handleChange('acme_staging', e.target.checked)}
                    className="mt-0.5 w-5 h-5 rounded border-amber-300 dark:border-amber-600 text-amber-600 focus:ring-amber-500 focus:ring-offset-0 bg-white dark:bg-slate-700"
                  />
                  <div>
                    <span className="text-sm font-semibold text-amber-800 dark:text-amber-400 flex items-center gap-2">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
                      </svg>
                      {t('system.acme.staging.label')}
                      <HelpTip contentKey="help.acme.staging" ns="settings" className="text-amber-800 dark:text-amber-400 hover:text-amber-900 dark:hover:text-amber-300" />
                    </span>
                    <p className="text-xs text-amber-700 dark:text-amber-500 mt-1" dangerouslySetInnerHTML={{ __html: t('system.acme.staging.description') }} />
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

                <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                  <label className="flex items-start gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={getValue('acme_auto_renew') ?? true}
                      onChange={(e) => handleChange('acme_auto_renew', e.target.checked)}
                      className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                    />
                    <div>
                      <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">{t('system.acme.autoRenew.enableLabel')}</span>
                      <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {t('system.acme.autoRenew.enableDescription')}
                      </p>
                    </div>
                  </label>
                </div>

                <div>
                  <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.acme.autoRenew.daysBeforeLabel')}</label>
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
                  className="px-4 py-2.5 bg-slate-100 hover:bg-slate-200 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg text-sm transition-colors flex items-center gap-2 disabled:opacity-50"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {testACMEMutation.isPending ? t('system.buttons.testing') : t('system.acme.testButton')}
                </button>
              </div>
            </div>
          )}

          {/* Bot Filter Defaults Settings */}
          {activeTab === 'botfilter' && (
            <div className="space-y-6">
              <div className="p-4 bg-slate-50 dark:bg-slate-700/30 border border-slate-200 dark:border-slate-700 rounded-lg">
                <div className="flex gap-4">
                  <div className="text-blue-600 dark:text-blue-400">
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-white flex items-center gap-2">
                      {t('system.botfilter.defaults.title')}
                      <HelpTip contentKey="help.botFilter.defaults" ns="settings" />
                    </h3>
                    <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                      {t('system.botfilter.defaults.description')}
                    </p>
                  </div>
                </div>
              </div>

              {/* Default Enable */}
              <div className="flex items-start gap-3 p-4 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-sm">
                <input
                  type="checkbox"
                  checked={getValue('bot_filter_default_enabled') ?? false}
                  onChange={(e) => handleChange('bot_filter_default_enabled', e.target.checked)}
                  className="mt-1 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 bg-white dark:bg-slate-700"
                />
                <div>
                  <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">{t('system.botfilter.enableDefault.label')}</span>
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                    {t('system.botfilter.enableDefault.description')}
                  </p>
                </div>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="space-y-6">
                  <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-slate-700 pb-2">{t('system.botfilter.options.title')}</h3>

                  <div className="space-y-4">
                    <label className="flex items-start gap-3 cursor-pointer group">
                      <input
                        type="checkbox"
                        checked={getValue('bot_filter_default_block_bad_bots') ?? true}
                        onChange={(e) => handleChange('bot_filter_default_block_bad_bots', e.target.checked)}
                        className="mt-0.5 w-4 h-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-blue-700 dark:group-hover:text-blue-400 transition-colors">{t('system.botfilter.options.blockBadBots.label')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('system.botfilter.options.blockBadBots.description')}</p>
                      </div>
                    </label>

                    <label className="flex items-start gap-3 cursor-pointer group">
                      <input
                        type="checkbox"
                        checked={getValue('bot_filter_default_block_ai_bots') ?? false}
                        onChange={(e) => handleChange('bot_filter_default_block_ai_bots', e.target.checked)}
                        className="mt-0.5 w-4 h-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-blue-700 dark:group-hover:text-blue-400 transition-colors">{t('system.botfilter.options.blockAiBots.label')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('system.botfilter.options.blockAiBots.description')}</p>
                      </div>
                    </label>

                    <label className="flex items-start gap-3 cursor-pointer group">
                      <input
                        type="checkbox"
                        checked={getValue('bot_filter_default_allow_search_engines') ?? true}
                        onChange={(e) => handleChange('bot_filter_default_allow_search_engines', e.target.checked)}
                        className="mt-0.5 w-4 h-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-blue-700 dark:group-hover:text-blue-400 transition-colors">{t('system.botfilter.options.allowSearchEngines.label')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('system.botfilter.options.allowSearchEngines.description')}</p>
                      </div>
                    </label>

                    <label className="flex items-start gap-3 cursor-pointer group">
                      <input
                        type="checkbox"
                        checked={getValue('bot_filter_default_challenge_suspicious') ?? false}
                        onChange={(e) => handleChange('bot_filter_default_challenge_suspicious', e.target.checked)}
                        className="mt-0.5 w-4 h-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-blue-700 dark:group-hover:text-blue-400 transition-colors">{t('system.botfilter.options.challengeSuspicious.label')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('system.botfilter.options.challengeSuspicious.description')}</p>
                      </div>
                    </label>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{t('system.botfilter.customAgents.label')}</label>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t('system.botfilter.customAgents.description')}</p>
                    <textarea
                      value={(editedSettings.bot_filter_default_custom_blocked_agents ?? settings?.bot_filter_default_custom_blocked_agents) || ''}
                      onChange={(e) => handleChange('bot_filter_default_custom_blocked_agents', e.target.value)}
                      className={`${inputClass} font-mono text-xs`}
                      rows={5}
                      placeholder={t('system.botfilter.customAgents.placeholder')}
                    />
                  </div>
                </div>

                <div className="space-y-6">
                  <div className="flex items-center justify-between border-b border-slate-200 dark:border-slate-700 pb-2">
                    <div className="flex items-center gap-2">
                      <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.botfilter.lists.title')}</h3>
                      <HelpTip contentKey="help.botFilter.lists" ns="settings" />
                      <span className="px-2 py-0.5 text-[10px] bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-full font-medium">{t('system.botfilter.lists.badge')}</span>
                    </div>
                  </div>
                  <p className="text-xs text-slate-500 dark:text-slate-400">
                    {t('system.botfilter.lists.description')}
                  </p>

                  <div>
                    <div className="flex justify-between items-baseline mb-2">
                      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">{t('system.botfilter.lists.badBots.label')}</label>
                      <span className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {t('system.botfilter.lists.patternCount', { count: (editedSettings.bot_list_bad_bots ?? settings?.bot_list_bad_bots)?.split('\n').filter((l: string) => l.trim()).length || 0 })}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t('system.botfilter.lists.badBots.description')}</p>
                    <textarea
                      value={(editedSettings.bot_list_bad_bots ?? settings?.bot_list_bad_bots) || ''}
                      onChange={(e) => handleChange('bot_list_bad_bots', e.target.value)}
                      className={`${inputClass} font-mono text-xs`}
                      rows={4}
                      placeholder={t('system.botfilter.lists.badBots.placeholder')}
                    />
                  </div>

                  <div>
                    <div className="flex justify-between items-baseline mb-2">
                      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">{t('system.botfilter.lists.aiBots.label')}</label>
                      <span className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {t('system.botfilter.lists.patternCount', { count: (editedSettings.bot_list_ai_bots ?? settings?.bot_list_ai_bots)?.split('\n').filter((l: string) => l.trim()).length || 0 })}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t('system.botfilter.lists.aiBots.description')}</p>
                    <textarea
                      value={(editedSettings.bot_list_ai_bots ?? settings?.bot_list_ai_bots) || ''}
                      onChange={(e) => handleChange('bot_list_ai_bots', e.target.value)}
                      className={`${inputClass} font-mono text-xs`}
                      rows={4}
                      placeholder={t('system.botfilter.lists.aiBots.placeholder')}
                    />
                  </div>

                  <div>
                    <div className="flex justify-between items-baseline mb-2">
                      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">{t('system.botfilter.lists.searchEngines.label')}</label>
                      <span className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                        {t('system.botfilter.lists.patternCount', { count: (editedSettings.bot_list_search_engines ?? settings?.bot_list_search_engines)?.split('\n').filter((l: string) => l.trim()).length || 0 })}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t('system.botfilter.lists.searchEngines.description')}</p>
                    <textarea
                      value={(editedSettings.bot_list_search_engines ?? settings?.bot_list_search_engines) || ''}
                      onChange={(e) => handleChange('bot_list_search_engines', e.target.value)}
                      className={`${inputClass} font-mono text-xs`}
                      rows={4}
                      placeholder={t('system.botfilter.lists.searchEngines.placeholder')}
                    />
                  </div>
                </div>
              </div>

              <div className="border-t border-slate-200 dark:border-slate-700 pt-6">
                <h3 className="font-medium text-slate-700 dark:text-slate-300 mb-3">{t('system.botfilter.summary.title')}</h3>
                <div className="flex gap-4">
                  <div className={`px-3 py-1.5 rounded-lg text-xs font-semibold ${(getValue('bot_filter_default_enabled') ?? false) ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400' : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400'
                    }`}>
                    Bot Filter: {(getValue('bot_filter_default_enabled') ?? false) ? t('system.botfilter.summary.active') : t('system.botfilter.summary.inactive')}
                  </div>
                  <div className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-red-50 dark:bg-red-900/10 text-red-700 dark:text-red-400">
                    {t('system.botfilter.summary.blocked')}: {
                      [
                        (getValue('bot_filter_default_block_bad_bots') ?? true) && t('system.botfilter.options.blockBadBots.label'),
                        (getValue('bot_filter_default_block_ai_bots') ?? false) && t('system.botfilter.options.blockAiBots.label')
                      ].filter(Boolean).join(', ') || '-'
                    }
                  </div>
                  <div className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-blue-50 dark:bg-blue-900/10 text-blue-700 dark:text-blue-400">
                    {t('system.botfilter.summary.allowed')}: {(getValue('bot_filter_default_allow_search_engines') ?? true) ? t('system.botfilter.options.allowSearchEngines.label') : '-'}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* WAF Auto-Ban Settings */}
          {activeTab === 'security' && (
            <div className="space-y-6">
              <div className={`p-5 rounded-xl border ${settings?.waf_auto_ban_enabled
                ? 'bg-emerald-50 dark:bg-emerald-900/10 border-emerald-200 dark:border-emerald-900/30'
                : 'bg-slate-50 dark:bg-slate-700/30 border-slate-200 dark:border-slate-700'
                }`}>
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-white">{t('system.waf.status.title')}</h3>
                    <p className="text-sm mt-1.5">
                      {settings?.waf_auto_ban_enabled ? (
                        <span className="text-emerald-700 dark:text-emerald-400">
                          {t('system.waf.status.activeDescription')}
                        </span>
                      ) : (
                        <span className="text-slate-600 dark:text-slate-400">
                          {t('system.waf.status.inactiveDescription')}
                        </span>
                      )}
                    </p>
                  </div>
                  <div className={`px-3 py-1 rounded-full text-xs font-semibold ${settings?.waf_auto_ban_enabled
                    ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                    : 'bg-slate-200 dark:bg-slate-600 text-slate-600 dark:text-slate-300'
                    }`}>
                    {settings?.waf_auto_ban_enabled ? t('system.waf.status.active') : t('system.waf.status.inactive')}
                  </div>
                </div>
              </div>

              <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                <label className="flex items-start gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={getValue('waf_auto_ban_enabled') ?? false}
                    onChange={(e) => handleChange('waf_auto_ban_enabled', e.target.checked)}
                    className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                  />
                  <div>
                    <span className="text-sm font-semibold text-slate-700 dark:text-slate-200 flex items-center gap-2">
                      {t('system.waf.enable.label')}
                      <HelpTip contentKey="help.waf.autoBan" ns="settings" />
                    </span>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                      {t('system.waf.enable.description')}
                    </p>
                  </div>
                </label>
              </div>

              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-slate-700 pb-2">{t('system.waf.config.title')}</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div>
                    <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.waf.config.threshold.label')}</label>
                    <input
                      type="number"
                      value={getValue('waf_auto_ban_threshold') ?? 10}
                      onChange={(e) => handleChange('waf_auto_ban_threshold', parseInt(e.target.value))}
                      className={inputClass}
                      min="1"
                    />
                    <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                      {t('system.waf.config.threshold.description')}
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.waf.config.window.label')}</label>
                    <input
                      type="number"
                      value={getValue('waf_auto_ban_window') ?? 300}
                      onChange={(e) => handleChange('waf_auto_ban_window', parseInt(e.target.value))}
                      className={inputClass}
                      min="30"
                    />
                    <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                      {t('system.waf.config.window.description')}
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.waf.config.duration.label')}</label>
                    <input
                      type="number"
                      value={getValue('waf_auto_ban_duration') ?? 3600}
                      onChange={(e) => handleChange('waf_auto_ban_duration', parseInt(e.target.value))}
                      className={inputClass}
                      min="0"
                    />
                    <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                      {t('system.waf.config.duration.description')}
                    </p>
                  </div>
                </div>
              </div>


              <div className="p-4 bg-blue-50 dark:bg-blue-900/10 rounded-lg border border-blue-100 dark:border-blue-900/20">
                <h4 className="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-1">{t('system.waf.info.title')}</h4>
                <p className="text-xs text-blue-700 dark:text-blue-400">
                  {t('system.waf.info.description')}
                </p>
              </div>
            </div>
          )}

          {/* System Logs Tab */}
          {activeTab === 'systemlogs' && (
            <div className="space-y-6">
              <div className="p-4 bg-slate-50 dark:bg-slate-700/30 border border-slate-200 dark:border-slate-700 rounded-lg">
                <div className="flex gap-4">
                  <div className="text-blue-600 dark:text-blue-400">
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="font-semibold text-slate-800 dark:text-white">
                      {t('system.systemlogs.title', 'System Log Collection')}
                    </h3>
                    <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                      {t('system.systemlogs.description', 'Configure which logs are collected from the system containers. Reducing log level can improve performance.')}
                    </p>
                  </div>
                </div>
              </div>

              {/* Enable/Disable */}
              <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                <label className="flex items-start gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={getLogConfigValue('enabled') ?? true}
                    onChange={(e) => setEditedLogConfig(prev => ({ ...prev, enabled: e.target.checked }))}
                    className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-600"
                  />
                  <div>
                    <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">
                      {t('system.systemlogs.enable.label', 'Enable System Log Collection')}
                    </span>
                    <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                      {t('system.systemlogs.enable.description', 'Collect logs from docker containers')}
                    </p>
                  </div>
                </label>
              </div>

              {/* Container Levels */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-slate-700 pb-2">
                  {t('system.systemlogs.levels.title', 'Container Log Levels')}
                </h3>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {['npg-proxy', 'npg-api', 'npg-db', 'npg-ui'].map((container) => {
                    const currentLevel = (editedLogConfig.levels?.[container]) || (systemLogConfig?.levels?.[container]) || 'info';
                    return (
                      <div key={container} className="p-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg">
                        <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                          {container}
                        </label>
                        <select
                          value={currentLevel}
                          onChange={(e) => {
                            const newLevels = {
                              ...(systemLogConfig?.levels || {}),
                              ...(editedLogConfig.levels || {}),
                              [container]: e.target.value
                            };
                            setEditedLogConfig(prev => ({ ...prev, levels: newLevels }));
                          }}
                          className={inputClass}
                        >
                          <option value="debug">Debug</option>
                          <option value="info">Info</option>
                          <option value="warn">Warning</option>
                          <option value="error">Error</option>
                          <option value="fatal">Fatal</option>
                        </select>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Exclude Patterns */}
              <div>
                <div className="flex justify-between items-baseline mb-2">
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
                    {t('system.systemlogs.exclude.label', 'Exclude Patterns')}
                  </label>
                  <span className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                    {t('system.systemlogs.exclude.count', { count: (getLogConfigValue('exclude_patterns') || []).length })}
                  </span>
                </div>
                <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">
                  {t('system.systemlogs.exclude.description', 'Log messages matching these patterns (regex) will be ignored. One pattern per line.')}
                </p>
                <textarea
                  value={excludePatternsText !== null
                    ? excludePatternsText
                    : (editedLogConfig.exclude_patterns !== undefined
                      ? editedLogConfig.exclude_patterns
                      : systemLogConfig?.exclude_patterns || []
                    ).join('\n')}
                  onChange={(e) => setExcludePatternsText(e.target.value)}
                  onBlur={(e) => {
                    const patterns = e.target.value.split('\n').filter(s => s.trim())
                    setEditedLogConfig(prev => ({ ...prev, exclude_patterns: patterns }))
                    setExcludePatternsText(null)
                  }}
                  className={`${inputClass} font-mono text-xs`}
                  rows={6}
                  placeholder="^/health
HEAD /"
                />
              </div>
            </div>
          )}


          {/* Maintenance Settings */}
          {
            activeTab === 'maintenance' && (
              <div className="space-y-6">
                {/* Log Retention Settings per Log Type */}
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <h3 className="text-base font-semibold text-slate-800 dark:text-white flex items-center gap-2">
                      {t('system.maintenance.logRetention.title')}
                      <HelpTip contentKey="help.maintenance.logRetention" ns="settings" />
                    </h3>
                  </div>
                  <p className="text-sm text-slate-500 dark:text-slate-400 mb-4">
                    {t('system.maintenance.logRetention.description')}
                  </p>

                  <div className="bg-slate-50 dark:bg-slate-700/30 rounded-xl border border-slate-200 dark:border-slate-700 divide-y divide-slate-200 dark:divide-slate-700">
                    {/* Access Logs */}
                    <div className="flex items-center justify-between p-4">
                      <div className="flex-1">
                        <label className="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-300">
                          <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                          {t('system.maintenance.logRetention.accessLogs.label')}
                        </label>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 ml-4">{t('system.maintenance.logRetention.accessLogs.description')}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="3650"
                          value={getValue('access_log_retention_days') ?? 1095}
                          onChange={(e) => handleChange('access_log_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400 w-8">{t('system.maintenance.logRetention.unit')}</span>
                        <span className="text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded w-16 text-center">
                          {t('system.maintenance.logRetention.years', { count: Math.round((getValue('access_log_retention_days') ?? 1095) / 365 * 10) / 10 })}
                        </span>
                      </div>
                    </div>

                    {/* WAF Events */}
                    <div className="flex items-center justify-between p-4">
                      <div className="flex-1">
                        <label className="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-300">
                          <div className="w-2 h-2 rounded-full bg-red-500"></div>
                          {t('system.maintenance.logRetention.wafEvents.label')}
                        </label>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 ml-4">{t('system.maintenance.logRetention.wafEvents.description')}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="3650"
                          value={getValue('waf_log_retention_days') ?? 90}
                          onChange={(e) => handleChange('waf_log_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400 w-8">{t('system.maintenance.logRetention.unit')}</span>
                        <span className="text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded w-16 text-center">
                          {t('system.maintenance.logRetention.months', { count: Math.round((getValue('waf_log_retention_days') ?? 90) / 30 * 10) / 10 })}
                        </span>
                      </div>
                    </div>

                    {/* Error Logs */}
                    <div className="flex items-center justify-between p-4">
                      <div className="flex-1">
                        <label className="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-300">
                          <div className="w-2 h-2 rounded-full bg-amber-500"></div>
                          {t('system.maintenance.logRetention.errorLogs.label')}
                        </label>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 ml-4">{t('system.maintenance.logRetention.errorLogs.description')}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="3650"
                          value={getValue('error_log_retention_days') ?? 30}
                          onChange={(e) => handleChange('error_log_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400 w-8">{t('system.maintenance.logRetention.unit')}</span>
                        <span className="text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded w-16 text-center">
                          {t('system.maintenance.logRetention.months', { count: Math.round((getValue('error_log_retention_days') ?? 30) / 30 * 10) / 10 })}
                        </span>
                      </div>
                    </div>

                    {/* System Logs */}
                    <div className="flex items-center justify-between p-4">
                      <div className="flex-1">
                        <label className="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-300">
                          <div className="w-2 h-2 rounded-full bg-slate-500"></div>
                          {t('system.maintenance.logRetention.systemLogs.label')}
                        </label>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 ml-4">{t('system.maintenance.logRetention.systemLogs.description')}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="3650"
                          value={getValue('system_log_retention_days') ?? 30}
                          onChange={(e) => handleChange('system_log_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400 w-8">{t('system.maintenance.logRetention.unit')}</span>
                        <span className="text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded w-16 text-center">
                          {t('system.maintenance.logRetention.months', { count: Math.round((getValue('system_log_retention_days') ?? 30) / 30 * 10) / 10 })}
                        </span>
                      </div>
                    </div>

                    {/* Admin Audit */}
                    <div className="flex items-center justify-between p-4">
                      <div className="flex-1">
                        <label className="flex items-center gap-2 text-sm font-semibold text-slate-700 dark:text-slate-300">
                          <div className="w-2 h-2 rounded-full bg-purple-500"></div>
                          {t('system.maintenance.logRetention.adminAudit.label')}
                        </label>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 ml-4">{t('system.maintenance.logRetention.adminAudit.description')}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="number"
                          min="1"
                          max="3650"
                          value={getValue('audit_log_retention_days') ?? 1095}
                          onChange={(e) => handleChange('audit_log_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400 w-8">{t('system.maintenance.logRetention.unit')}</span>
                        <span className="text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded w-16 text-center">
                          {t('system.maintenance.logRetention.years', { count: Math.round((getValue('audit_log_retention_days') ?? 1095) / 365 * 10) / 10 })}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Quick Presets */}
                  <div className="mt-4 flex items-center gap-3 flex-wrap">
                    <span className="text-sm font-medium text-slate-500 dark:text-slate-400">{t('system.maintenance.presets.title')}:</span>
                    <button
                      type="button"
                      onClick={() => {
                        handleChange('access_log_retention_days', 1095);
                        handleChange('waf_log_retention_days', 90);
                        handleChange('error_log_retention_days', 30);
                        handleChange('system_log_retention_days', 30);
                        handleChange('audit_log_retention_days', 1095);
                      }}
                      className="px-3 py-1.5 text-sm font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/10 hover:bg-blue-100 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                    >
                      {t('system.maintenance.presets.default')}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        handleChange('access_log_retention_days', 365);
                        handleChange('waf_log_retention_days', 180);
                        handleChange('error_log_retention_days', 90);
                        handleChange('system_log_retention_days', 90);
                        handleChange('audit_log_retention_days', 1825);
                      }}
                      className="px-3 py-1.5 text-sm font-medium text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-900/10 hover:bg-emerald-100 dark:hover:bg-emerald-900/20 rounded-lg transition-colors"
                    >
                      {t('system.maintenance.presets.extended')}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        handleChange('access_log_retention_days', 30);
                        handleChange('waf_log_retention_days', 30);
                        handleChange('error_log_retention_days', 7);
                        handleChange('system_log_retention_days', 7);
                        handleChange('audit_log_retention_days', 365);
                      }}
                      className="px-3 py-1.5 text-sm font-medium text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/10 hover:bg-amber-100 dark:hover:bg-amber-900/20 rounded-lg transition-colors"
                    >
                      {t('system.maintenance.presets.minimal')}
                    </button>
                  </div>
                </div>

                {/* Stats & Backup Retention */}
                <div className="border-t border-slate-200 dark:border-slate-700 pt-6">
                  <div className="flex items-center gap-2 mb-4">
                    <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
                    </svg>
                    <h3 className="text-base font-semibold text-slate-800 dark:text-white">{t('system.maintenance.otherRetention.title')}</h3>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="bg-slate-50 dark:bg-slate-700/30 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                      <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.otherRetention.stats.label')}</label>
                      <div className="flex items-center gap-2 mt-2">
                        <input
                          type="number"
                          min="1"
                          max="365"
                          value={getValue('stats_retention_days') ?? 90}
                          onChange={(e) => handleChange('stats_retention_days', parseInt(e.target.value))}
                          className="w-24 px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400">{t('system.maintenance.logRetention.unit')}</span>
                      </div>
                      <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                        {t('system.maintenance.otherRetention.stats.description')}
                      </p>
                    </div>

                    <div className="bg-slate-50 dark:bg-slate-700/30 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                      <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.otherRetention.backup.label')}</label>
                      <div className="flex items-center gap-2 mt-2">
                        <input
                          type="number"
                          min="1"
                          max="100"
                          value={getValue('backup_retention_count') ?? 10}
                          onChange={(e) => handleChange('backup_retention_count', parseInt(e.target.value))}
                          className="w-24 px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                        <span className="text-sm text-slate-500 dark:text-slate-400">{t('system.maintenance.otherRetention.backup.unit')}</span>
                      </div>
                      <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                        {t('system.maintenance.otherRetention.backup.description')}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Auto Backup */}
                <div className="border-t border-slate-200 dark:border-slate-700 pt-6">
                  <div className="flex items-center gap-2 mb-4">
                    <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2" />
                    </svg>
                    <h3 className="text-base font-semibold text-slate-800 dark:text-white">{t('system.maintenance.autoBackup.title')}</h3>
                  </div>

                  <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700">
                    <label className="flex items-start gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={getValue('auto_backup_enabled') ?? false}
                        onChange={(e) => handleChange('auto_backup_enabled', e.target.checked)}
                        className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">{t('system.maintenance.autoBackup.enableLabel')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                          {t('system.maintenance.autoBackup.enableDescription')}
                        </p>
                      </div>
                    </label>
                  </div>

                  <div className="mt-5">
                    <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.autoBackup.scheduleLabel')}</label>
                    <input
                      type="text"
                      value={(editedSettings.auto_backup_schedule ?? settings?.auto_backup_schedule) || '0 2 * * *'}
                      onChange={(e) => handleChange('auto_backup_schedule', e.target.value)}
                      className="mt-1 w-full md:w-64 px-3 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg font-mono text-sm text-slate-700 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="0 2 * * *"
                    />
                    <p className="mt-2 text-xs text-slate-500 dark:text-slate-400" dangerouslySetInnerHTML={{ __html: t('system.maintenance.autoBackup.scheduleHelp') }} />
                  </div>
                </div>

                {/* Raw Log Files */}
                <div className="border-t border-slate-200 dark:border-slate-700 pt-6">
                  <div className="flex items-center gap-2 mb-2">
                    <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                    </svg>
                    <h3 className="text-base font-semibold text-slate-800 dark:text-white flex items-center gap-2">
                      {t('system.maintenance.rawLogs.title')}
                      <HelpTip contentKey="help.maintenance.rawLogs" ns="settings" />
                    </h3>
                  </div>
                  <p className="text-sm text-slate-500 dark:text-slate-400 mb-4">
                    {t('system.maintenance.rawLogs.description')}
                  </p>

                  <div className="py-3 px-4 bg-slate-50 dark:bg-slate-700/30 rounded-lg border border-slate-200 dark:border-slate-700 mb-5">
                    <label className="flex items-start gap-3 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={getValue('raw_log_enabled') ?? false}
                        onChange={(e) => handleChange('raw_log_enabled', e.target.checked)}
                        className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-700"
                      />
                      <div>
                        <span className="text-sm font-semibold text-slate-700 dark:text-slate-200">{t('system.maintenance.rawLogs.enable.label')}</span>
                        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1" dangerouslySetInnerHTML={{ __html: t('system.maintenance.rawLogs.enable.description') }} />
                      </div>
                    </label>
                  </div>

                  {getValue('raw_log_enabled') && (
                    <div className="space-y-5 pl-2 border-l-2 border-blue-200 dark:border-blue-900/30 ml-2">
                      {/* Log Rotation Settings */}
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                          <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.rawLogs.maxSize.label')}</label>
                          <div className="flex items-center gap-2 mt-2">
                            <input
                              type="number"
                              min="1"
                              max="1000"
                              value={getValue('raw_log_max_size_mb') ?? 100}
                              onChange={(e) => handleChange('raw_log_max_size_mb', parseInt(e.target.value))}
                              className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <span className="text-sm text-slate-500 dark:text-slate-400">MB</span>
                          </div>
                          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                            {t('system.maintenance.rawLogs.maxSize.help')}
                          </p>
                        </div>

                        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                          <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.rawLogs.rotateCount.label')}</label>
                          <div className="flex items-center gap-2 mt-2">
                            <input
                              type="number"
                              min="1"
                              max="100"
                              value={getValue('raw_log_rotate_count') ?? 5}
                              onChange={(e) => handleChange('raw_log_rotate_count', parseInt(e.target.value))}
                              className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <span className="text-sm text-slate-500 dark:text-slate-400">{t('system.maintenance.rawLogs.rotateCount.unit')}</span>
                          </div>
                          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                            {t('system.maintenance.rawLogs.rotateCount.help')}
                          </p>
                        </div>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                          <label className="block text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.rawLogs.retention.label')}</label>
                          <div className="flex items-center gap-2 mt-2">
                            <input
                              type="number"
                              min="1"
                              max="365"
                              value={getValue('raw_log_retention_days') ?? 7}
                              onChange={(e) => handleChange('raw_log_retention_days', parseInt(e.target.value))}
                              className="w-24 px-3 py-2 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-sm text-slate-700 dark:text-white text-right focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <span className="text-sm text-slate-500 dark:text-slate-400">{t('system.maintenance.logRetention.unit')}</span>
                          </div>
                          <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
                            {t('system.maintenance.rawLogs.retention.help')}
                          </p>
                        </div>

                        <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
                          <div className="py-1">
                            <label className="flex items-start gap-3 cursor-pointer">
                              <input
                                type="checkbox"
                                checked={getValue('raw_log_compress_rotated') ?? true}
                                onChange={(e) => handleChange('raw_log_compress_rotated', e.target.checked)}
                                className="mt-0.5 w-5 h-5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-slate-700"
                              />
                              <div>
                                <span className="text-sm font-semibold text-slate-700 dark:text-slate-300">{t('system.maintenance.rawLogs.compress.label')}</span>
                                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                                  {t('system.maintenance.rawLogs.compress.help')}
                                </p>
                              </div>
                            </label>
                          </div>
                        </div>
                      </div>

                      {/* Storage Estimate */}
                      <div className="p-4 bg-blue-50 dark:bg-blue-900/10 rounded-lg border border-blue-200 dark:border-blue-900/20">
                        <div className="flex items-start gap-3">
                          <svg className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                          <div className="text-sm text-blue-800 dark:text-blue-300">
                            <p className="font-medium">{t('system.maintenance.rawLogs.storage.title')}</p>
                            <p className="mt-1">
                              {t('system.maintenance.rawLogs.storage.estimate', {
                                size: ((getValue('raw_log_max_size_mb') ?? 100) * (getValue('raw_log_rotate_count') ?? 5) * 2 * (getValue('raw_log_compress_rotated') ? 0.1 : 1)).toFixed(0)
                              })}
                              {getValue('raw_log_compress_rotated') && ` (${t('system.maintenance.rawLogs.storage.compressed')})`}
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )
          }

          {/* Log Files Management */}
          {
            activeTab === 'logfiles' && (
              <div className="space-y-6">
                {/* Status */}
                <div className={`p-5 rounded-xl border ${logFilesData?.raw_log_enabled
                  ? 'bg-emerald-50 dark:bg-emerald-900/10 border-emerald-200 dark:border-emerald-900/30'
                  : 'bg-slate-50 dark:bg-slate-700/30 border-slate-200 dark:border-slate-700'
                  }`}>
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold text-slate-800 dark:text-white">{t('system.logfiles.status.title')}</h3>
                      <p className="text-sm mt-1.5">
                        {logFilesData?.raw_log_enabled ? (
                          <span className="text-emerald-700 dark:text-emerald-400">
                            {t('system.logfiles.status.activeDescription', {
                              count: logFilesData?.total_count ?? 0,
                              size: formatBytes(logFilesData?.total_size ?? 0)
                            })}
                          </span>
                        ) : (
                          <span className="text-slate-600 dark:text-slate-400">
                            {t('system.logfiles.status.inactiveDescription')}
                          </span>
                        )}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => refetchLogFiles()}
                        className="px-3 py-2 text-sm font-medium bg-white dark:bg-slate-700 text-slate-700 dark:text-white border border-slate-300 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                        </svg>
                      </button>
                      {logFilesData?.raw_log_enabled && (
                        <button
                          onClick={() => rotateMutation.mutate()}
                          disabled={rotateMutation.isPending}
                          className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
                        >
                          {rotateMutation.isPending ? t('system.logfiles.actions.rotating') : t('system.logfiles.actions.rotate')}
                        </button>
                      )}
                    </div>
                  </div>
                </div>

                {/* File List */}
                {(logFilesData?.files?.length ?? 0) > 0 ? (
                  <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                    <div className="px-4 py-3 bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-700">
                      <h3 className="font-semibold text-slate-700 dark:text-slate-300 text-sm">{t('system.logfiles.list.title')}</h3>
                    </div>
                    <div className="divide-y divide-slate-100 dark:divide-slate-700">
                      {logFilesData?.files.map((file: LogFileInfo) => (
                        <div key={file.name} className="flex items-center justify-between p-4 hover:bg-slate-50 dark:hover:bg-slate-700/50">
                          <div className="flex items-center gap-3">
                            <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${file.log_type === 'access' ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' :
                              file.log_type === 'error' ? 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400' :
                                'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400'
                              }`}>
                              {file.is_compressed ? (
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                                </svg>
                              ) : (
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                              )}
                            </div>
                            <div>
                              <div className="font-medium text-slate-800 dark:text-white text-sm">{file.name}</div>
                              <div className="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-2 mt-0.5">
                                <span>{formatBytes(file.size)}</span>
                                <span>•</span>
                                <span>{new Date(file.modified_at).toLocaleString(i18n.language === 'ko' ? 'ko-KR' : 'en-US')}</span>
                                {file.is_compressed && (
                                  <>
                                    <span>•</span>
                                    <span className="text-amber-600 dark:text-amber-400">{t('system.logfiles.list.compressed')}</span>
                                  </>
                                )}
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleViewFile(file.name)}
                              className="p-2 text-slate-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                              title={t('system.logfiles.actions.preview')}
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                              </svg>
                            </button>
                            <button
                              onClick={() => handleDownloadFile(file.name)}
                              className="p-2 text-slate-400 dark:text-slate-500 hover:text-emerald-600 dark:hover:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 rounded-lg transition-colors"
                              title={t('system.logfiles.actions.download')}
                            >
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                              </svg>
                            </button>
                            {file.name !== 'access.log' && file.name !== 'error.log' && (
                              <button
                                onClick={() => setConfirmDelete(file.name)}
                                className="p-2 text-slate-400 dark:text-slate-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                                title={t('system.logfiles.actions.delete')}
                              >
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                              </button>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-12 bg-slate-50 dark:bg-slate-800/50 rounded-xl border border-slate-200 dark:border-slate-700">
                    <svg className="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <p className="text-slate-500 dark:text-slate-400">{t('system.logfiles.list.empty')}</p>
                    <p className="text-sm text-slate-400 dark:text-slate-500 mt-1">
                      {t('system.logfiles.list.emptyHint')}
                    </p>
                  </div>
                )}

                {/* File Viewer Modal */}
                {viewingFile && (
                  <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
                    <div className="bg-white dark:bg-slate-800 rounded-xl shadow-2xl w-full max-w-4xl max-h-[80vh] flex flex-col">
                      <div className="px-5 py-4 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between">
                        <h3 className="font-semibold text-slate-800 dark:text-white">{viewingFile}</h3>
                        <button
                          onClick={() => {
                            setViewingFile(null);
                            setViewContent('');
                          }}
                          className="p-2 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                        >
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                        </button>
                      </div>
                      <div className="flex-1 overflow-auto p-4">
                        {viewFileMutation.isPending ? (
                          <div className="flex items-center justify-center py-12">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                          </div>
                        ) : (
                          <pre className="text-xs text-slate-700 dark:text-slate-300 font-mono whitespace-pre-wrap bg-slate-50 dark:bg-slate-900 p-4 rounded-lg overflow-auto">
                            {viewContent || t('system.logfiles.list.emptyFile')}
                          </pre>
                        )}
                      </div>
                      <div className="px-5 py-3 border-t border-slate-200 dark:border-slate-700 flex justify-end gap-2">
                        <button
                          onClick={() => handleDownloadFile(viewingFile)}
                          className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                        >
                          {t('system.logfiles.actions.download')}
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {/* Delete Confirmation Modal */}
                {confirmDelete && (
                  <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
                    <div className="bg-white dark:bg-slate-800 rounded-xl shadow-2xl w-full max-w-md p-6">
                      <div className="flex items-center gap-4 mb-4">
                        <div className="p-3 bg-red-100 dark:bg-red-900/30 rounded-full">
                          <svg className="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </div>
                        <div>
                          <h3 className="font-semibold text-slate-800 dark:text-white">{t('system.logfiles.delete.title')}</h3>
                          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
                            <Trans
                              i18nKey="system.logfiles.delete.confirm"
                              values={{ file: confirmDelete }}
                              components={{ 1: <span className="font-medium text-slate-700 dark:text-white" /> }}
                            />
                          </p>
                        </div>
                      </div>
                      <div className="flex justify-end gap-3">
                        <button
                          onClick={() => setConfirmDelete(null)}
                          className="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 rounded-lg transition-colors"
                        >
                          {t('system.buttons.cancel')}
                        </button>
                        <button
                          onClick={() => deleteFileMutation.mutate(confirmDelete)}
                          disabled={deleteFileMutation.isPending}
                          className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors disabled:opacity-50"
                        >
                          {deleteFileMutation.isPending ? t('system.logfiles.delete.deleting') : t('system.logfiles.actions.delete')}
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )
          }
        </div >
      </div >
    </div >
  );
}
