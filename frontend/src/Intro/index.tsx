import { Button, ButtonGroup, Callout, Card, Divider, H5 } from '@blueprintjs/core';
import i18next from 'i18next';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import './index.css';
import { IsNoSelfUpdate } from '../../wailsjs/go/app/App';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import BlueSkyLogo from '../assets/icons/bluesky.svg';
import DiscordIcon from '../assets/icons/discord.svg';
import GithubIcon from '../assets/icons/github.svg';
import OpenCollectiveIcon from '../assets/icons/opencollective.svg';
import RedditIcon from '../assets/icons/reddit.svg';
import { ThemeType, useTheme } from '../common/ThemeManager';
import { AppHeader } from '../components/AppHeader';
import { SUPPORTED_LANGUAGES } from '../constants/languages';
import { SOCIAL_LINKS } from '../constants/urls';
import { useProxyState } from '../context/ProxyStateContext';
import { FilterLists } from '../FilterLists';
import { FilterListType } from '../FilterLists/types';
import { AutostartSwitch } from '../SettingsManager/AutostartSwitch';
import { UpdatePolicyRadioGroup } from '../SettingsManager/UpdatePolicyRadioGroup';
import { StartStopButton } from '../StartStopButton';

import { LanguageList } from './LanguageList';

interface IntroOverlayProps {
  isOpen: boolean;
  onClose: () => void;
}

function RegionalFilterLists() {
  return <FilterLists initialType={FilterListType.REGIONAL} hideTypeSelector />;
}

export function IntroOverlay({ isOpen, onClose }: IntroOverlayProps) {
  const { t } = useTranslation();

  const [currentScreen, setCurrentScreen] = useState(4);
  const [isLanguageTransitioning, setIsLanguageTransitioning] = useState(false);
  const [animatedLanguage, setAnimatedLanguage] = useState('en-US');
  const [welcomeText, setWelcomeText] = useState(t('introOverlay.welcome'));
  const [descriptionText, setDescriptionText] = useState(t('introOverlay.screen1.description'));
  const { proxyState } = useProxyState();

  const { effectiveTheme } = useTheme();

  const TOTAL_SCREENS = 4;

  const getTranslationsFor = (languageCode: string) => {
    const tfixed = i18next.getFixedT(languageCode);
    return {
      welcome: tfixed('introOverlay.welcome'),
      description: tfixed('introOverlay.screen1.description'),
    };
  };

  useEffect(() => {
    if (isOpen) setCurrentScreen(1);
  }, [isOpen]);

  useEffect(() => {
    if (currentScreen !== 1 || !isOpen) {
      return () => {};
    }

    let transitionTimeout: ReturnType<typeof setTimeout>;

    const animationInterval = setInterval(() => {
      setIsLanguageTransitioning(true);

      transitionTimeout = setTimeout(() => {
        const currentIndex = SUPPORTED_LANGUAGES.findIndex((lang) => lang.value === animatedLanguage);
        const nextIndex = (currentIndex + 1) % SUPPORTED_LANGUAGES.length;
        const nextLanguage = SUPPORTED_LANGUAGES[nextIndex].value;

        setAnimatedLanguage(nextLanguage);
        const texts = getTranslationsFor(nextLanguage);
        setWelcomeText(texts.welcome);
        setDescriptionText(texts.description);

        setTimeout(() => setIsLanguageTransitioning(false), 50);
      }, 300);
    }, 4000);

    return () => {
      clearInterval(animationInterval);
      clearTimeout(transitionTimeout);
    };
  }, [currentScreen, isOpen, animatedLanguage]);

  const completeIntro = () => {
    localStorage.setItem('zen-intro-completed', 'true');
    onClose();
  };

  useEffect(() => {
    if (currentScreen === TOTAL_SCREENS && proxyState === 'on') {
      completeIntro();
    }
  }, [proxyState, currentScreen]);

  const handleNextScreen = () => {
    if (currentScreen < TOTAL_SCREENS) setCurrentScreen(currentScreen + 1);
  };

  const screens = [
    // Screen 1: Welcome & Language Selection
    <div className="intro-screen" key="screen-1">
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
      <LanguageList />
    </div>,

    // Screen 2: Filter Lists
    <div className="intro-screen" key="screen-2">
      <h2 className="intro-heading">{t('introOverlay.screen2.title')}</h2>
      <p>{t('introOverlay.screen2.description')}</p>
      <p>{t('introOverlay.screen2.recommendation')}</p>
      <div className="filter-lists">
        <RegionalFilterLists />
      </div>
    </div>,

    // Screen 3: Settings
    <div className="intro-screen" key="screen-3">
      <h2 className="intro-heading">{t('introOverlay.screen3.title')}</h2>
      <p>{t('introOverlay.screen3.description')}</p>

      <Card elevation={1} className="settings-card">
        <AutostartSwitch />

        {!IsNoSelfUpdate() && (
          <>
            <Divider className="settings-divider" />
            <UpdatePolicyRadioGroup />
          </>
        )}
      </Card>

      <Callout icon="info-sign" intent="primary" className="settings-note">
        {t('introOverlay.screen3.settingsNote')}
      </Callout>
    </div>,

    // Screen 4: Social Links & Donation
    <div className="intro-screen" key="screen-4">
      <h2 className="intro-heading">{t('introOverlay.screen4.title')}</h2>
      <p>{t('introOverlay.screen4.description')}</p>

      <Card elevation={1} className="connect-card">
        <H5>{t('introOverlay.screen4.socialText')}</H5>

        <div className="social-links-grid">
          <div className="social-row">
            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.GITHUB)} className="social-button">
              <img src={GithubIcon} className="social-icon" alt="GitHub" />
              GitHub
            </Button>

            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.BLUESKY)} className="social-button">
              <img src={BlueSkyLogo} className="social-icon" alt="Bluesky" />
              Bluesky
            </Button>
          </div>

          <div className="social-row">
            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.REDDIT)} className="social-button">
              <img src={RedditIcon} className="social-icon" alt="Reddit" />
              Reddit
            </Button>

            <Button fill onClick={() => BrowserOpenURL(SOCIAL_LINKS.DISCORD)} className="social-button">
              <img src={DiscordIcon} className="social-icon" alt="Discord" />
              Discord
            </Button>
          </div>
        </div>

        <Divider className="section-divider" />

        <p>{t('introOverlay.screen4.donateText')}</p>
        <Button
          icon={<img src={OpenCollectiveIcon} className="social-icon" alt="Open Collective" />}
          onClick={() => BrowserOpenURL(SOCIAL_LINKS.OPEN_COLLECTIVE)}
        >
          Open Collective
        </Button>
      </Card>
    </div>,
  ];

  if (!isOpen) return null;

  return (
    <div className={`intro-fullscreen${effectiveTheme === ThemeType.DARK ? ' bp5-dark' : ''}`}>
      <AppHeader />
      <div className="intro-main-content">{screens[currentScreen - 1]}</div>
      <div className="intro-footer">
        {currentScreen < TOTAL_SCREENS ? (
          <ButtonGroup fill size="large">
            <Button fill variant="outlined" onClick={completeIntro} className="skip-button">
              {t('introOverlay.buttons.skip')}
            </Button>
            <Button fill intent="primary" onClick={handleNextScreen} endIcon="arrow-right">
              {t('introOverlay.buttons.next')}
            </Button>
          </ButtonGroup>
        ) : (
          <StartStopButton />
        )}
      </div>
    </div>
  );
}
