import { Button, ButtonGroup } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import './index.css';
import { ThemeType, useTheme } from '../common/ThemeManager';
import { AppHeader } from '../components/AppHeader';
import { useProxyState } from '../context/ProxyStateContext';
import { StartStopButton } from '../StartStopButton';

import { ConnectScreen } from './ConnectScreen';
import { FilterListsScreen } from './FilterListsScreen';
import { SettingsScreen } from './SettingsScreen';
import { WelcomeScreen } from './WelcomeScreen';

interface IntroOverlayProps {
  isOpen: boolean;
  onClose: () => void;
}

const TOTAL_SCREENS = 4;

export function IntroOverlay({ isOpen, onClose }: IntroOverlayProps) {
  const { t } = useTranslation();

  const [currentScreen, setCurrentScreen] = useState(1);
  const { proxyState } = useProxyState();

  const { effectiveTheme } = useTheme();

  useEffect(() => {
    if (isOpen) setCurrentScreen(1);
  }, [isOpen]);

  useEffect(() => {
    if (currentScreen === TOTAL_SCREENS && proxyState === 'on') {
      onClose();
    }
  }, [proxyState, currentScreen]);

  const handleNextScreen = () => {
    if (currentScreen < TOTAL_SCREENS) setCurrentScreen(currentScreen + 1);
  };

  const screens = [
    <WelcomeScreen key="screen-1" />,
    <FilterListsScreen key="screen-2" />,
    <SettingsScreen key="screen-3" />,
    <ConnectScreen key="screen-4" />,
  ];

  if (!isOpen) return null;

  return (
    <div className={`intro-fullscreen${effectiveTheme === ThemeType.DARK ? ' bp5-dark' : ''}`}>
      <AppHeader />
      <div className="intro-main-content">{screens[currentScreen - 1]}</div>
      <div className="intro-footer">
        {currentScreen < TOTAL_SCREENS ? (
          <ButtonGroup fill size="large">
            <Button fill variant="outlined" onClick={onClose} className="skip-button">
              {t('introOverlay.buttons.skip')}
            </Button>
            <Button fill intent="primary" onClick={handleNextScreen} endIcon="arrow-right">
              {t('introOverlay.buttons.next')}
            </Button>
            <Button
              fill
              intent="primary"
              onClick={() => {
                setCurrentScreen(currentScreen - 1);
              }}
              endIcon="arrow-right"
            >
              Back
            </Button>
          </ButtonGroup>
        ) : (
          <StartStopButton />
        )}
      </div>
    </div>
  );
}
