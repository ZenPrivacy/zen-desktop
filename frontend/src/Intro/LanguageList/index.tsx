import { Card, Radio } from '@blueprintjs/core';

import { SUPPORTED_LANGUAGES } from '../../constants/languages';

import './index.css';

interface LanguageListProps {
  selectedLanguage: string;
  onSelect: (lang: string) => void;
}

export function LanguageList({ selectedLanguage, onSelect }: LanguageListProps) {
  return (
    <div className="language-list">
      {SUPPORTED_LANGUAGES.map((language) => (
        <Card
          key={language.value}
          className={`language-option ${selectedLanguage === language.value ? 'selected' : ''}`}
          interactive
          elevation={selectedLanguage === language.value ? 2 : 0}
          onClick={() => onSelect(language.value)}
        >
          <div className="language-content">
            <Radio
              checked={selectedLanguage === language.value}
              onChange={() => onSelect(language.value)}
              className="language-radio"
              label={language.label}
            />
          </div>
        </Card>
      ))}
    </div>
  );
}
