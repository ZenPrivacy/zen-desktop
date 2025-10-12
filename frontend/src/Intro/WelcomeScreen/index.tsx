import i18next from 'i18next';
import { useEffect, useState } from 'react';

import { changeLocale, getCurrentLocale } from '../../i18n';

import { LocaleList } from './LocaleList';

import './index.css';

const getTranslationsFor = (languageCode: string) => {
  const tfixed = i18next.getFixedT(languageCode);
  return {
    welcome: tfixed('intro.welcome.title'),
    description: tfixed('intro.welcome.description'),
  };
};

export function WelcomeScreen() {
  const [locale, setLocale] = useState(getCurrentLocale);
  const [welcomeText, setWelcomeText] = useState('');
  const [descriptionText, setDescriptionText] = useState('');

  useEffect(() => {
    if (!locale) return;

    const texts = getTranslationsFor(locale);
    setWelcomeText(texts.welcome);
    setDescriptionText(texts.description);
  }, [locale]);

  return (
    <div className="intro-screen">
      <div>
        <h2 className="welcome-slide bp5-heading intro-heading" key={`welcome-${locale}`}>
          👋 {welcomeText}
        </h2>
        <p className="welcome-slide intro-description" key={`desc-${locale}`}>
          {descriptionText}
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
