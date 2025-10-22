import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import { GetLocale, SetLocale } from '../../wailsjs/go/cfg/Config';

import deDE from './locales/de-DE.json';
import enUS from './locales/en-US.json';
import frFR from './locales/fr-FR.json';
import itIT from './locales/it-IT.json';
import kkKZ from './locales/kk-KZ.json';
import ruRU from './locales/ru-RU.json';
import zhCN from './locales/zh-CN.json';
import zhTW from './locales/zh-TW.json';

export const SUPPORTED_LOCALES = [
  'en',
  'en-US',
  'de',
  'de-DE',
  'it',
  'it-IT',
  'kk',
  'kk-KZ',
  'ru',
  'ru-RU',
  'zh',
  'zh-CN',
  'zh-TW',
  'fr',
  'fr-FR',
] as const;
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];
export const FALLBACK_LOCALE: SupportedLocale = 'en-US';

export interface LocaleItem {
  value: SupportedLocale;
  label: string;
}

export const LOCALE_LABELS: LocaleItem[] = [
  { value: 'en-US', label: 'English' },
  { value: 'de-DE', label: 'Deutsch' },
  { value: 'kk-KZ', label: 'Қазақша' },
  { value: 'ru-RU', label: 'Русский' },
  { value: 'zh-CN', label: '中文（简体）' },
  { value: 'zh-TW', label: '中文（繁體）' },
  { value: 'it-IT', label: 'Italiano' },
  { value: 'fr-FR', label: 'Français' },
];

export function detectSystemLocale(): SupportedLocale {
  const browserLang = navigator.language;
  const detected = SUPPORTED_LOCALES.includes(browserLang as any) ? (browserLang as SupportedLocale) : FALLBACK_LOCALE;

  return detected;
}

export function getCurrentLocale(): SupportedLocale {
  return (i18n.language as SupportedLocale) || FALLBACK_LOCALE;
}

export async function changeLocale(locale: SupportedLocale) {
  const normalized = SUPPORTED_LOCALES.includes(locale) ? locale : FALLBACK_LOCALE;
  await i18n.changeLanguage(normalized);
  await SetLocale(normalized);
}

export async function initI18n() {
  let locale = await GetLocale();
  if (locale === '') {
    const detected = detectSystemLocale();
    await SetLocale(detected);
    locale = detected;
  }

  return i18n.use(initReactI18next).init({
    resources: {
      en: { translation: enUS },
      'en-US': { translation: enUS },
      de: { translation: deDE },
      'de-DE': { translation: deDE },
      kk: { translation: kkKZ },
      'kk-KZ': { translation: kkKZ },
      ru: { translation: ruRU },
      'ru-RU': { translation: ruRU },
      zh: { translation: zhCN },
      'zh-CN': { translation: zhCN },
      'zh-TW': { translation: zhTW },
      it: { translation: itIT },
      'it-IT': { translation: itIT },
      fr: { translation: frFR },
      'fr-FR': { translation: frFR },
    },
    lng: locale,
    fallbackLng: FALLBACK_LOCALE,
    returnNull: false,
    returnEmptyString: false,
    interpolation: {
      escapeValue: false,
    },
    react: {
      useSuspense: false,
    },
  });
}
