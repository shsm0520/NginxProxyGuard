import type { Log } from '../types';

interface GeoIPBadgeProps {
  log: Log;
}

export function GeoIPBadge({ log }: GeoIPBadgeProps) {
  if (!log.geo_country_code) return <span className="text-slate-400">-</span>;

  return (
    <span
      className="px-2 py-0.5 bg-slate-100 dark:bg-slate-700 rounded text-xs font-medium text-slate-700 dark:text-slate-300"
      title={`${log.geo_country || log.geo_country_code} ${log.geo_org ? `(${log.geo_org})` : ''}`}
    >
      {log.geo_country_code}
    </span>
  );
}
