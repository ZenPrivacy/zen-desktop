import i18next, { changeLanguage } from 'i18next';
import { useEffect, useState } from 'react';

import { SUPPORTED_LANGUAGES } from '../../constants/languages';
import { LanguageList } from '../LanguageList';

const getTranslationsFor = (languageCode: string) => {
  const tfixed = i18next.getFixedT(languageCode);
  return {
    welcome: tfixed('introOverlay.welcome'),
    description: tfixed('introOverlay.screen1.description'),
  };
};

export function WelcomeScreen() {
  const [isLanguageTransitioning, setIsLanguageTransitioning] = useState(false);
  const [animatedLanguage, setAnimatedLanguage] = useState<string | null>(null);
  const [welcomeText, setWelcomeText] = useState('');
  const [descriptionText, setDescriptionText] = useState('');

  useEffect(() => {
    let langIndex = 0;
    let transitionTimeout: ReturnType<typeof setTimeout>;

    const setTexts = (language: string) => {
      const texts = getTranslationsFor(language);
      setWelcomeText(texts.welcome);
      setDescriptionText(texts.description);
    };

    setTexts(SUPPORTED_LANGUAGES[langIndex].value);

    const animationInterval = setInterval(() => {
      setIsLanguageTransitioning(true);

      transitionTimeout = setTimeout(() => {
        langIndex = (langIndex + 1) % SUPPORTED_LANGUAGES.length;
        setTexts(SUPPORTED_LANGUAGES[langIndex].value);
        setIsLanguageTransitioning(false);
      }, 300);
    }, 4000);

    return () => {
      clearInterval(animationInterval);
      clearTimeout(transitionTimeout);
    };
  }, []);

  useEffect(() => {
    if (!animatedLanguage) return;

    const texts = getTranslationsFor(animatedLanguage);
    setWelcomeText(texts.welcome);
    setDescriptionText(texts.description);
    setIsLanguageTransitioning(false);
  }, [animatedLanguage]);

  const handleLanguageSelect = (lang: string) => {
    setAnimatedLanguage(lang);
    changeLanguage(lang);
  };

  return (
    <div className="intro-screen">
      <div className="text-animation-container">
        <h2
          className={`intro-heading animated-welcome ${
            isLanguageTransitioning ? 'language-fade-out' : 'language-fade-in'
          }`}
          key={`welcome-${animatedLanguage}`}
        >
          {welcomeText}
        </h2>
        <p
          className={`animated-description ${isLanguageTransitioning ? 'language-fade-out' : 'language-fade-in'}`}
          key={`desc-${animatedLanguage}`}
        >
          {descriptionText}
        </p>
      </div>
      <LanguageList onSelect={handleLanguageSelect} selectedLanguage={animatedLanguage ?? ''} />
    </div>
  );
}
