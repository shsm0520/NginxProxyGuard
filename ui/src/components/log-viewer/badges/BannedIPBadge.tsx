import { useTranslation } from 'react-i18next';

interface BannedIPBadgeProps {
  ip?: string;
  isBanned?: boolean;
}

export function BannedIPBadge({ ip, isBanned }: BannedIPBadgeProps) {
  const { t } = useTranslation('logs');

  if (!ip) return <span className="text-slate-400">-</span>;

  const ipClassName = 'font-mono text-xs sm:text-sm leading-snug break-all';

  if (isBanned) {
    return (
      <span className="flex flex-wrap items-center gap-1.5 min-w-0" title={ip}>
        <span className={`${ipClassName} text-red-600 dark:text-red-400 min-w-0`}>{ip}</span>
        <span className="shrink-0 px-1.5 py-0.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 text-xs font-medium rounded">
          {t('badges.banned', { defaultValue: 'Banned' })}
        </span>
      </span>
    );
  }

  return (
    <span className={`block ${ipClassName} text-slate-600 dark:text-slate-300`} title={ip}>
      {ip}
    </span>
  );
}
