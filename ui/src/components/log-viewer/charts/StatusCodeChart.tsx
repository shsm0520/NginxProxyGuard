import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import type { StatusCodeChartProps } from '../types';

export const StatusCodeChart = memo(function StatusCodeChart({ data }: StatusCodeChartProps) {
  const { t } = useTranslation('logs');
  const total = data.reduce((sum, d) => sum + d.count, 0);

  const getStatusColor = (code: number) => {
    if (code >= 500) return { bg: 'bg-red-500', text: 'text-red-600' };
    if (code >= 400) return { bg: 'bg-yellow-500', text: 'text-yellow-600' };
    if (code >= 300) return { bg: 'bg-blue-500', text: 'text-blue-600' };
    return { bg: 'bg-green-500', text: 'text-green-600' };
  };

  // Calculate cumulative percentages for the donut
  let cumulative = 0;
  const segments = data.map(d => {
    const percentage = (d.count / total) * 100;
    const segment = {
      ...d,
      percentage,
      startOffset: cumulative,
      color: getStatusColor(d.status_code),
    };
    cumulative += percentage;
    return segment;
  });

  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <h4 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">{t('stats.statusCodes')}</h4>
      <div className="flex items-center gap-4">
        {/* Simple donut representation using stacked bars */}
        <div className="relative w-24 h-24">
          <svg viewBox="0 0 36 36" className="w-24 h-24 transform -rotate-90">
            {segments.map((seg, i) => (
              <circle
                key={i}
                cx="18"
                cy="18"
                r="15.915"
                fill="transparent"
                stroke="currentColor"
                strokeWidth="3"
                strokeDasharray={`${seg.percentage} ${100 - seg.percentage}`}
                strokeDashoffset={-seg.startOffset}
                className={seg.color.text}
              />
            ))}
          </svg>
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-xs font-bold text-slate-700 dark:text-slate-200">{total.toLocaleString()}</span>
          </div>
        </div>

        {/* Legend */}
        <div className="flex-1 space-y-1">
          {segments.slice(0, 6).map((seg, i) => (
            <div key={i} className="flex items-center gap-2 text-xs">
              <span className={`w-2 h-2 rounded-full ${seg.color.bg}`} />
              <span className="text-slate-600 dark:text-slate-400">{seg.status_code}</span>
              <span className="text-slate-400 dark:text-slate-500 ml-auto">{seg.percentage.toFixed(1)}%</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
});
