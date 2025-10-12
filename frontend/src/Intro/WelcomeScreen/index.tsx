import i18next from 'i18next';
import { useEffect, useRef, useState } from 'react';

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
  const [transition, setTransition] = useState(false);
  const [locale, setLocale] = useState(getCurrentLocale);
  const [welcomeText, setWelcomeText] = useState('');
  const [descriptionText, setDescriptionText] = useState('');

  useEffect(() => {
    if (!locale) return;

    const texts = getTranslationsFor(locale);
    setWelcomeText(texts.welcome);
    setDescriptionText(texts.description);
    setTransition(false);
  }, [locale]);

  const transitionTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (transitionTimeoutRef.current) {
        clearTimeout(transitionTimeoutRef.current);
      }
    };
  }, []);

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
          if (transitionTimeoutRef.current) {
            clearTimeout(transitionTimeoutRef.current);
          }
          transitionTimeoutRef.current = window.setTimeout(() => {
            setTransition(false);
            transitionTimeoutRef.current = null;
          }, 300);
        }}
        selectedLocale={locale}
      />
    </div>
  );
}
