import { Card, Radio } from '@blueprintjs/core';

import { SUPPORTED_LANGUAGES } from '../../../constants/languages';
import { SupportedLocale } from '../../../i18n';

import './index.css';

interface LocaleListProps {
  selectedLocale: string;
  onSelect: (lang: SupportedLocale) => void;
}

export function LocaleList({ selectedLocale, onSelect }: LocaleListProps) {
  return (
    <div className="locale-list">
      {SUPPORTED_LANGUAGES.map((locale) => (
        <Card
          key={locale.value}
          className={`locale-option ${selectedLocale === locale.value ? 'selected' : ''}`}
          interactive
          elevation={selectedLocale === locale.value ? 2 : 0}
          onClick={() => onSelect(locale.value)}
        >
          <div className="locale-content">
            <Radio
              checked={selectedLocale === locale.value}
              onChange={() => onSelect(locale.value)}
              className="locale-radio"
              label={locale.label}
            />
          </div>
        </Card>
      ))}
    </div>
  );
}
