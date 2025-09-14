import { useState } from 'react';
import { Card, Radio } from '@blueprintjs/core';
import { changeLocale, getCurrentLocale, SupportedLocale } from '../../i18n';
import { SUPPORTED_LANGUAGES } from '../../constants/languages';

import './index.css';

export function LanguageList() {
  const [selectedLanguage, setSelectedLanguage] = useState(getCurrentLocale());

  const handleLanguageSelect = (languageCode: SupportedLocale) => {
    setSelectedLanguage(languageCode);
    changeLocale(languageCode);
  };

  return (
    <div className="language-list">
      {SUPPORTED_LANGUAGES.map((language) => (
        <Card
          key={language.value}
          className={`language-option ${selectedLanguage === language.value ? 'selected' : ''}`}
          interactive
          elevation={selectedLanguage === language.value ? 2 : 0}
          onClick={() => handleLanguageSelect(language.value)}
        >
          <div className="language-content">
            <Radio
              checked={selectedLanguage === language.value}
              onChange={() => handleLanguageSelect(language.value)}
              className="language-radio"
            />
            <div className="language-text">
              <div>{language.label}</div>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
