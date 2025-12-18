interface CategoryIconProps {
  name: string;
}

const iconMap: Record<string, string> = {
  XSS: 'text-purple-600 dark:text-purple-400 bg-purple-100 dark:bg-purple-900/30',
  SQLI: 'text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30',
  RCE: 'text-orange-600 dark:text-orange-400 bg-orange-100 dark:bg-orange-900/30',
  LFI: 'text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900/30',
  RFI: 'text-amber-600 dark:text-amber-400 bg-amber-100 dark:bg-amber-900/30',
  PHP: 'text-indigo-600 dark:text-indigo-400 bg-indigo-100 dark:bg-indigo-900/30',
  JAVA: 'text-blue-600 dark:text-blue-400 bg-blue-100 dark:bg-blue-900/30',
  PROTOCOL: 'text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800',
  SCANNER: 'text-pink-600 dark:text-pink-400 bg-pink-100 dark:bg-pink-900/30',
  SESSION: 'text-cyan-600 dark:text-cyan-400 bg-cyan-100 dark:bg-cyan-900/30',
};

export function CategoryIcon({ name }: CategoryIconProps) {
  const colorClass = iconMap[name] || 'text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800';

  return (
    <div className={`w-8 h-8 rounded flex items-center justify-center text-xs font-bold ${colorClass}`}>
      {name.slice(0, 3)}
    </div>
  );
}
