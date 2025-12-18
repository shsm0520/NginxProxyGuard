import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import { getGeoRestriction, setGeoRestriction, deleteGeoRestriction, getCountryCodes } from '../api/access';
import type { CreateGeoRestrictionRequest } from '../types/access';

interface GeoRestrictionPanelProps {
  proxyHostId: string;
}

export default function GeoRestrictionPanel({ proxyHostId }: GeoRestrictionPanelProps) {
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [mode, setMode] = useState<'whitelist' | 'blacklist'>('blacklist');
  const [selectedCountries, setSelectedCountries] = useState<string[]>([]);
  const [enabled, setEnabled] = useState(true);
  const [challengeMode, setChallengeMode] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState('');

  const { data: geoRestriction, isLoading: isLoadingGeo } = useQuery({
    queryKey: ['geo-restriction', proxyHostId],
    queryFn: () => getGeoRestriction(proxyHostId).catch(() => null),
  });

  const { data: countryCodes, isLoading: isLoadingCountries } = useQuery({
    queryKey: ['country-codes'],
    queryFn: getCountryCodes,
  });

  useEffect(() => {
    if (geoRestriction) {
      setMode(geoRestriction.mode);
      setSelectedCountries(geoRestriction.countries || []);
      setEnabled(geoRestriction.enabled);
      setChallengeMode(geoRestriction.challenge_mode || false);
    }
  }, [geoRestriction]);

  const saveMutation = useMutation({
    mutationFn: (data: CreateGeoRestrictionRequest) => setGeoRestriction(proxyHostId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['geo-restriction', proxyHostId] });
      setIsEditing(false);
      setError('');
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to save geo restriction');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteGeoRestriction(proxyHostId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['geo-restriction', proxyHostId] });
      setSelectedCountries([]);
      setMode('blacklist');
      setEnabled(true);
      setChallengeMode(false);
      setError('');
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to delete geo restriction');
    },
  });

  const handleSave = () => {
    if (selectedCountries.length === 0) {
      setError('Please select at least one country');
      return;
    }
    saveMutation.mutate({
      mode,
      countries: selectedCountries,
      enabled,
      challenge_mode: challengeMode,
    });
  };

  const handleDelete = () => {
    if (confirm('Are you sure you want to remove geo restriction?')) {
      deleteMutation.mutate();
    }
  };

  const toggleCountry = (code: string) => {
    setSelectedCountries(prev =>
      prev.includes(code)
        ? prev.filter(c => c !== code)
        : [...prev, code]
    );
  };

  const filteredCountries = countryCodes
    ? Object.entries(countryCodes).filter(([code, name]) =>
        code.toLowerCase().includes(searchTerm.toLowerCase()) ||
        name.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : [];

  if (isLoadingGeo || isLoadingCountries) {
    return (
      <div className="flex justify-center items-center h-32">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  const hasRestriction = geoRestriction && geoRestriction.countries?.length > 0;

  return (
    <div className="bg-white rounded-lg border p-4">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-medium text-gray-900">Geo Restriction</h3>
        {!isEditing && (
          <div className="space-x-2">
            <button
              onClick={() => setIsEditing(true)}
              className="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              {hasRestriction ? 'Edit' : 'Configure'}
            </button>
            {hasRestriction && (
              <button
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="px-3 py-1.5 text-sm text-red-600 border border-red-300 rounded hover:bg-red-50"
              >
                Remove
              </button>
            )}
          </div>
        )}
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded mb-4 text-sm">
          {error}
        </div>
      )}

      {!isEditing ? (
        hasRestriction ? (
          <div className="space-y-3">
            <div className="flex gap-4 text-sm">
              <span className="text-gray-500">Mode:</span>
              <span className={`font-medium ${geoRestriction.mode === 'whitelist' ? 'text-green-700' : 'text-red-700'}`}>
                {geoRestriction.mode === 'whitelist' ? 'Allow only' : 'Block'} selected countries
              </span>
            </div>
            <div className="flex gap-4 text-sm">
              <span className="text-gray-500">Status:</span>
              <span className={`font-medium ${geoRestriction.enabled ? 'text-green-700' : 'text-gray-500'}`}>
                {geoRestriction.enabled ? 'Active' : 'Disabled'}
              </span>
            </div>
            <div className="flex gap-4 text-sm">
              <span className="text-gray-500">Action:</span>
              <span className={`font-medium ${geoRestriction.challenge_mode ? 'text-blue-700' : 'text-red-700'}`}>
                {geoRestriction.challenge_mode ? 'Show CAPTCHA Challenge' : 'Block (403)'}
              </span>
            </div>
            <div>
              <span className="text-sm text-gray-500">Countries ({geoRestriction.countries?.length || 0}):</span>
              <div className="flex flex-wrap gap-1 mt-2">
                {geoRestriction.countries?.map(code => (
                  <span key={code} className="px-2 py-0.5 text-xs bg-gray-100 text-gray-700 rounded">
                    {code} - {countryCodes?.[code] || code}
                  </span>
                ))}
              </div>
            </div>
          </div>
        ) : (
          <p className="text-sm text-gray-500">
            No geo restriction configured. Click "Configure" to block or allow traffic from specific countries.
          </p>
        )
      ) : (
        <div className="space-y-4">
          <div className="flex gap-4">
            <label className="flex items-center gap-2">
              <input
                type="radio"
                name="mode"
                checked={mode === 'blacklist'}
                onChange={() => setMode('blacklist')}
                className="text-indigo-600 focus:ring-indigo-500"
              />
              <span className="text-sm text-gray-700">Block selected countries</span>
            </label>
            <label className="flex items-center gap-2">
              <input
                type="radio"
                name="mode"
                checked={mode === 'whitelist'}
                onChange={() => setMode('whitelist')}
                className="text-indigo-600 focus:ring-indigo-500"
              />
              <span className="text-sm text-gray-700">Allow only selected countries</span>
            </label>
          </div>

          <div className="flex flex-col gap-2">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="text-sm text-gray-700">Enabled</span>
            </label>

            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={challengeMode}
                onChange={(e) => setChallengeMode(e.target.checked)}
                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span className="text-sm text-gray-700">
                Challenge Mode (show CAPTCHA instead of blocking)
              </span>
            </label>
            {challengeMode && (
              <p className="text-xs text-gray-500 ml-6">
                Users from restricted countries will see a CAPTCHA challenge.
                After passing, they can access the site for a configurable time period.
                Configure CAPTCHA settings in Global Settings.
              </p>
            )}
          </div>

          <div>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Search countries..."
            />
          </div>

          {selectedCountries.length > 0 && (
            <div>
              <span className="text-sm text-gray-500">Selected ({selectedCountries.length}):</span>
              <div className="flex flex-wrap gap-1 mt-1">
                {selectedCountries.map(code => (
                  <button
                    key={code}
                    onClick={() => toggleCountry(code)}
                    className="px-2 py-0.5 text-xs bg-indigo-100 text-indigo-700 rounded hover:bg-indigo-200"
                  >
                    {code} x
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="max-h-48 overflow-y-auto border rounded-md p-2">
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-1">
              {filteredCountries.map(([code, name]) => (
                <label
                  key={code}
                  className={`flex items-center gap-2 p-1.5 rounded cursor-pointer text-sm ${
                    selectedCountries.includes(code) ? 'bg-indigo-50' : 'hover:bg-gray-50'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={selectedCountries.includes(code)}
                    onChange={() => toggleCountry(code)}
                    className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                  />
                  <span className="truncate">{code} - {name}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t">
            <button
              onClick={() => {
                setIsEditing(false);
                setError('');
                if (geoRestriction) {
                  setMode(geoRestriction.mode);
                  setSelectedCountries(geoRestriction.countries || []);
                  setEnabled(geoRestriction.enabled);
                  setChallengeMode(geoRestriction.challenge_mode || false);
                }
              }}
              className="px-3 py-1.5 text-sm text-gray-700 bg-gray-100 rounded hover:bg-gray-200"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={saveMutation.isPending}
              className="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
            >
              {saveMutation.isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
