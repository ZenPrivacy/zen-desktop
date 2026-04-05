import { defineConfig } from 'i18next-cli';

export default defineConfig({
  locales: ['en-US', 'de-DE', 'kk-KZ', 'ru-RU', 'it-IT', 'tr-TR', 'zh-TW', 'zh-CN', 'fr-FR'],
  extract: {
    input: ['src/**/*.{js,jsx,ts,tsx}'],
    output: 'src/i18n/locales/{{language}}.json',
    defaultNS: 'translation',
    keySeparator: '.',
    nsSeparator: false,
    functions: ['t', '*.t'],
    transComponents: ['Trans'],
    sort: true,
  },
});
