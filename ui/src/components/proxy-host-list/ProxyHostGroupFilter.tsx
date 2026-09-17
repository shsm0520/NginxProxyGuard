import { useTranslation } from 'react-i18next'
import type { ProxyHostGroups, ProxyHostListFilter } from '../../types/proxy-host'
import { XIcon } from '../common/listui'

interface ProxyHostGroupFilterProps {
  groups?: ProxyHostGroups
  filter: ProxyHostListFilter
  onChange: (next: ProxyHostListFilter) => void
}

const SELECT =
  'px-2 py-1 text-sm border border-slate-200 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent'

// The panel is drawn from GET /proxy-hosts/groups, so it only ever offers
// buckets that actually match something. Tag chips toggle (AND semantics, the
// same rule the API applies); the three selects are single-valued.
export function ProxyHostGroupFilter({ groups, filter, onChange }: ProxyHostGroupFilterProps) {
  const { t } = useTranslation('proxyHost')
  const active = filter.tags ?? []
  const hasAny = active.length > 0 || !!filter.domain || !!filter.upstream || filter.enabled !== undefined
  if (!groups) return null
  const toggleTag = (tag: string) =>
    onChange({ ...filter, tags: active.includes(tag) ? active.filter((x) => x !== tag) : [...active, tag] })

  return (
    <div className="mb-4 flex flex-col gap-2" data-testid="group-filter">
      <div className="flex flex-wrap items-center gap-2">
        {groups.tags.length > 0 && (
          <div className="flex flex-wrap items-center gap-1.5">
            <span className="text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400">
              {t('list.groups.tags')}
            </span>
            {groups.tags.map((g) => {
              const on = active.includes(g.name)
              return (
                <button
                  key={g.name}
                  type="button"
                  onClick={() => toggleTag(g.name)}
                  aria-pressed={on}
                  data-testid={`group-tag-${g.name}`}
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors ${
                    on
                      ? 'bg-primary-600 text-white'
                      : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-slate-700 dark:text-slate-200 dark:hover:bg-slate-600'
                  }`}
                >
                  {g.name} <span className="opacity-70">{g.count}</span>
                </button>
              )
            })}
          </div>
        )}
        <select
          value={filter.domain ?? ''}
          onChange={(e) => onChange({ ...filter, domain: e.target.value || undefined })}
          className={SELECT}
          aria-label={t('list.groups.domain')}
          data-testid="group-domain"
        >
          <option value="">
            {t('list.groups.domain')}: {t('list.groups.any')}
          </option>
          {groups.domains.map((g) => (
            <option key={g.name} value={g.name}>
              {g.name} ({g.count})
            </option>
          ))}
        </select>
        <select
          value={filter.upstream ?? ''}
          onChange={(e) => onChange({ ...filter, upstream: e.target.value || undefined })}
          className={SELECT}
          aria-label={t('list.groups.upstream')}
          data-testid="group-upstream"
        >
          <option value="">
            {t('list.groups.upstream')}: {t('list.groups.any')}
          </option>
          {groups.upstreams.map((g) => (
            <option key={g.name} value={g.name}>
              {g.name} ({g.count})
            </option>
          ))}
        </select>
        <select
          value={filter.enabled === undefined ? '' : String(filter.enabled)}
          onChange={(e) =>
            onChange({ ...filter, enabled: e.target.value === '' ? undefined : e.target.value === 'true' })
          }
          className={SELECT}
          aria-label={t('list.groups.status')}
          data-testid="group-status"
        >
          <option value="">
            {t('list.groups.status')}: {t('list.groups.any')}
          </option>
          <option value="true">
            {t('list.filter.enabled')} ({groups.status.enabled})
          </option>
          <option value="false">
            {t('list.filter.disabled')} ({groups.status.disabled})
          </option>
        </select>
        {hasAny && (
          <button
            type="button"
            onClick={() => onChange({})}
            className="inline-flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200"
            data-testid="group-clear"
          >
            <XIcon /> {t('list.groups.clear')}
          </button>
        )}
      </div>
    </div>
  )
}
