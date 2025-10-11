import i18next from 'i18next';
import { useEffect, useState } from 'react';

import { LOCALE_LABELS, changeLocale, getCurrentLocale } from '../../i18n';

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
  const [transition, setTransition] = useState(false);
  const [locale, setLocale] = useState<string | null>(null);
  const [welcomeText, setWelcomeText] = useState('');
  const [descriptionText, setDescriptionText] = useState('');

  useEffect(() => {
    setLocale((LOCALE_LABELS.find((item) => item.value === getCurrentLocale()) || LOCALE_LABELS[0]).value);
  }, []);

  useEffect(() => {
    if (!locale) return;

    const texts = getTranslationsFor(locale);
    setWelcomeText(texts.welcome);
    setDescriptionText(texts.description);
    setTransition(false);
  }, [locale]);

  return (
    <div className="intro-screen">
      <div>
        <h2
          className={`animated-welcome bp5-heading intro-heading ${
            transition ? 'welcome-fade-out' : 'welcome-fade-in'
          }`}
          key={`welcome-${locale}`}
        >
          👋 {welcomeText}
        </h2>
        <p
          className={`animated-description intro-description ${transition ? 'welcome-fade-out' : 'welcome-fade-in'}`}
          key={`desc-${locale}`}
        >
          {descriptionText}
        </p>
      </div>
      <LocaleList
        onSelect={(locale) => {
          setTransition(true);
          setLocale(locale);
          changeLocale(locale);
          setTimeout(() => {
            setTransition(false);
          }, 300);
        }}
        selectedLocale={locale ?? ''}
      />
    </div>
  );
}
