import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import type { BarChartProps } from '../types';

export const BarChart = memo(function BarChart({ data, title, maxItems = 10 }: BarChartProps) {
  const { t } = useTranslation('logs');
  const displayData = data.slice(0, maxItems);
  const maxValue = Math.max(...displayData.map(d => d.value), 1);

  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <h4 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">{title}</h4>
      <div className="space-y-2">
        {displayData.length === 0 ? (
          <p className="text-xs text-slate-400 dark:text-slate-500 text-center py-4">{t('charts.noData')}</p>
        ) : (
          displayData.map((item, index) => (
            <div key={index} className="flex items-center gap-2">
              <span className="text-xs text-slate-600 dark:text-slate-400 w-36 truncate font-mono" title={item.label}>
                {item.label}
              </span>
              <div className="flex-1 h-5 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden min-w-[60px]">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${item.color || 'bg-primary-500'}`}
                  style={{ width: `${(item.value / maxValue) * 100}%` }}
                />
              </div>
              <span className="text-xs font-medium text-slate-700 dark:text-slate-300 w-14 text-right">
                {item.value.toLocaleString()}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
});
