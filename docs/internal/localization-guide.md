# Localization guide

This guide will help you contribute translations to Zen.

## Overview of the localization system

- Zen uses the [i18next](https://www.i18next.com/) framework for localization, as well as [react-i18next](https://react.i18next.com/) for easy integration with React.
- Translation files are stored as JSON files in the [`frontend/src/i18n/locales`](/frontend/src/i18n/locales) directory.
- The main language is English, and the default locale is `en-US`.
- Internally, Zen uses locales rather than languages. [Learn more about the difference](https://poeditor.com/blog/locale-vs-language/).

## Currently supported locales

Currently supported locales can be found in the `SUPPORTED_LOCALES` array in the [`frontend/src/i18n/index.ts`](/frontend/src/i18n/index.ts#L9) file.

## Adding a new locale

To add a new locale to Zen, you'll need to:

1. Fork the repository
2. Set up your development environment
3. Create a new translation file
4. Update the i18n configuration
5. Update the language selector
6. Test your translations
7. Submit a Pull Request

### Step 1: Fork the repository

Follow [these instructions](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo) to create a fork of Zen.

### Step 2: Set up your development environment

1. Install Go, Node.js with npm, Wails, and optionally Task as described in [requirements.md](requirements.md)
2. In the project root directory, run `wails dev` (or `task` if you have Task installed) to start the development server

### Step 3: Create a translation file

1. Copy the English template at [`en-US.json`](/frontend/src/i18n/locales/en-US.json) to a new file named according to your locale code (e.g., `de-DE.json` for German)
2. Translate all strings in the new file, keeping the JSON structure intact

### Step 4: Update the i18n configuration

You'll need to modify two files:

1. In [`frontend/i18next-parser.config.js`](/frontend/i18next-parser.config.js#L9), add your locale code to the `locales` array:

  ```typescript
  locales: ['en-US', 'ru-RU', 'your-locale-code'],
  ```

2. In [`frontend/src/i18n/index.ts`](/frontend/src/i18n/index.ts):

- Add an [import for your translation file](/frontend/src/i18n/index.ts#L6).
- Add your locale to the [`SUPPORTED_LOCALES` array](/frontend/src/i18n/index.ts#L9).
- Add your translation to the [`resources`](/frontend/src/i18n/index.ts#L39) object in the `initI18n` function. Since a user's preferred locale [may not include a region code](https://developer.mozilla.org/en-US/docs/Web/API/Navigator/language#value), you should also provide a fallback for the base language (e.g., `de` for `de-DE`).

  ```typescript
  resources: {
    en: { translation: enUS },
    'en-US': { translation: enUS },
    ru: { translation: ruRU },
    'ru-RU': { translation: ruRU },
    de: { translation: deDE },
    'de-DE': { translation: deDE },
  },
  ```

### Step 5: Update the language selector

Update the [`LocaleSelector` component](/frontend/src/SettingsManager/LocaleSelector/index.tsx#L12) to include your locale in the dropdown menu.

### Step 6: Test your translations

1. Navigate to Settings > Language to switch to your locale
2. Test the application thoroughly with your locale enabled

### Step 7: Submit a Pull Request

Follow [this guide](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request-from-a-fork) to submit a Pull Request with your changes.

> [!IMPORTANT]
> By contributing translations, you agree that your work will be licensed under the MIT License as [used by the project](/LICENSE).

## Translation best practices

1. **Maintain variables**: Keep special placeholders like `{{error}}` intact
2. **Preserve HTML**: If the original string contains HTML tags, maintain them in your translation
3. **Test your translations**: Run the application to verify your translations appear correctly
4. **Keep similar length**: Try to keep translations similar in length to the original

## Extracting translation keys

If you're a developer adding new text to the application, you can run:

```sh
task frontend:extract-translations
```

This will scan the source code and update the translation files with new keys.

## Need help?

If you have any questions about the translation process, feel free to ask in the [Discussions](https://github.com/anfragment/zen/discussions/categories/contributor-q-a) section of the GitHub repository, join our [Discord server](https://discord.gg/jSzEwby7JY), or contact one of the project leads directly via email.
