import { Card, Radio } from '@blueprintjs/core';

import { LOCALE_LABELS, SupportedLocale } from '../../../i18n';

import './index.css';

interface LocaleListProps {
  selectedLocale: SupportedLocale;
  onSelect: (locale: SupportedLocale) => void;
}

export function LocaleList({ selectedLocale, onSelect }: LocaleListProps) {
  return (
    <div className="locale-list">
      {LOCALE_LABELS.map((locale) => (
        <Card
          key={locale.value}
          className={`locale-option ${selectedLocale === locale.value ? 'selected' : ''}`}
          interactive
          elevation={selectedLocale === locale.value ? 2 : 0}
          onClick={() => onSelect(locale.value)}
        >
          <div className="locale-content">
            <Radio checked={selectedLocale === locale.value} className="locale-radio" label={locale.label} />
          </div>
        </Card>
      ))}
    </div>
  );
}
