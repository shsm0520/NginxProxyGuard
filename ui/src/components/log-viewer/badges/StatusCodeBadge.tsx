interface StatusCodeBadgeProps {
  code: number;
}

export function StatusCodeBadge({ code }: StatusCodeBadgeProps) {
  let color = 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
  if (code >= 200 && code < 300) color = 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
  else if (code >= 300 && code < 400) color = 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
  else if (code >= 400 && code < 500) color = 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
  else if (code >= 500) color = 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
  else if (code === 101) color = 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400';

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${color}`}>
      {code}
    </span>
  );
}
