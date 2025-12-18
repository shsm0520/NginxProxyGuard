import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { STORAGE_KEY, SUPPORTED_LANGUAGES, LANGUAGE_NAMES, type SupportedLanguage } from '../index'
import { getLanguage, setLanguage as apiSetLanguage } from '../../api/auth'

export function useLanguage(isAuthenticated: boolean = false) {
  const { i18n } = useTranslation()
  const queryClient = useQueryClient()

  // Fetch current language from server (only for logged-in users)
  const { data: serverLanguage } = useQuery({
    queryKey: ['language'],
    queryFn: getLanguage,
    staleTime: Infinity,
    retry: false,
    enabled: isAuthenticated,
  })

  // Current language (from i18n)
  const currentLanguage = i18n.language as SupportedLanguage

  // Mutation to update language
  const mutation = useMutation({
    mutationFn: async (language: SupportedLanguage) => {
      // Update localStorage immediately
      localStorage.setItem(STORAGE_KEY, language)

      // Update i18n
      await i18n.changeLanguage(language)

      // Try to sync with server (may fail if not logged in)
      try {
        await apiSetLanguage(language)
      } catch {
        // Ignore server errors - localStorage will persist the preference
      }

      return language
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['language'] })
    },
  })

  const changeLanguage = (language: SupportedLanguage) => {
    mutation.mutate(language)
  }

  // Sync with server language on load
  const syncWithServer = async () => {
    if (serverLanguage && serverLanguage !== currentLanguage) {
      await i18n.changeLanguage(serverLanguage)
      localStorage.setItem(STORAGE_KEY, serverLanguage)
    }
  }

  return {
    currentLanguage,
    changeLanguage,
    isChanging: mutation.isPending,
    supportedLanguages: SUPPORTED_LANGUAGES,
    languageNames: LANGUAGE_NAMES,
    syncWithServer,
  }
}
