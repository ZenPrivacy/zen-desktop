import { getFixedT } from 'i18next';
import { useState } from 'react';

import { changeLocale, getCurrentLocale } from '@/i18n';

import { LocaleList } from './LocaleList';

import './index.css';

export function WelcomeScreen() {
  const [locale, setLocale] = useState(getCurrentLocale);

  const t = getFixedT(locale);

  return (
    <div className="intro-screen">
      <div>
        <h2 className="welcome-slide bp6-heading intro-heading" key={`welcome-${locale}`}>
          👋 {t('intro.welcome.title')}
        </h2>
        <p className="welcome-slide intro-description" key={`desc-${locale}`}>
          {t('intro.welcome.description')}
        </p>
      </div>
      <LocaleList
        onSelect={(locale) => {
          setLocale(locale);
          changeLocale(locale);
        }}
        selectedLocale={locale}
      />
    </div>
  );
}
