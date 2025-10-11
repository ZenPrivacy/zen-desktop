import { SupportedLocale } from '../i18n';

export interface LocaleItem {
  value: SupportedLocale;
  label: string;
}

export const SUPPORTED_LANGUAGES: LocaleItem[] = [
  { value: 'en-US', label: 'English' },
  { value: 'de-DE', label: 'Deutsch' },
  { value: 'kk-KZ', label: 'Қазақша' },
  { value: 'ru-RU', label: 'Русский' },
  { value: 'zh-CN', label: '中文（简体)' },
  { value: 'it-IT', label: 'Italiano' },
];

export const DEFAULT_LOCALE: SupportedLocale = 'en-US';
