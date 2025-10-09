import i18next from 'i18next';
import { useEffect, useState } from 'react';

import { SUPPORTED_LANGUAGES } from '../../constants/languages';
import { changeLocale, getCurrentLocale } from '../../i18n';

import { LocaleList } from './LocaleList';

import './index.css';

const getTranslationsFor = (languageCode: string) => {
  const tfixed = i18next.getFixedT(languageCode);
  return {
    welcome: tfixed('introOverlay.welcome'),
    description: tfixed('introOverlay.screen1.description'),
  };
};

export function WelcomeScreen() {
  const [transition, setTransition] = useState(false);
  const [locale, setLocale] = useState<string | null>(null);
  const [welcomeText, setWelcomeText] = useState('');
  const [descriptionText, setDescriptionText] = useState('');

  useEffect(() => {
    setLocale((SUPPORTED_LANGUAGES.find((item) => item.value === getCurrentLocale()) || SUPPORTED_LANGUAGES[0]).value);
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
      <div className="text-animation-container">
        <h2
          className={`animated-welcome bp5-heading ${transition ? 'welcome-fade-out' : 'welcome-fade-in'}`}
          key={`welcome-${locale}`}
        >
          👋 {welcomeText}
        </h2>
        <p
          className={`animated-description ${transition ? 'welcome-fade-out' : 'welcome-fade-in'}`}
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
