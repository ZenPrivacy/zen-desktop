import { Button, ButtonGroup, ProgressBar } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import './index.css';
import { GetFilterListsByLocales } from '../../wailsjs/go/cfg/Config';
import { cfg } from '../../wailsjs/go/models';
import { ThemeType, useTheme } from '../common/ThemeManager';
import { AppHeader } from '../components/AppHeader';
import { useProxyState } from '../context/ProxyStateContext';
import { StartStopButton } from '../StartStopButton';

import { ConnectScreen } from './ConnectScreen';
import { FilterListsScreen } from './FilterListsScreen';
import { SettingsScreen } from './SettingsScreen';
import { WelcomeScreen } from './WelcomeScreen';

interface IntroOverlayProps {
  onClose: () => void;
}

export function IntroOverlay({ onClose }: IntroOverlayProps) {
  const { t } = useTranslation();

  const [currentScreen, setCurrentScreen] = useState(1);
  const [filterLists, setFilterLists] = useState<cfg.FilterList[]>([]);

  useEffect(() => {
    GetFilterListsByLocales(navigator.languages as string[]).then((filterLists) => {
      if (filterLists) setFilterLists(filterLists);
    });
  }, []);

  const { proxyState } = useProxyState();
  const { effectiveTheme } = useTheme();

  const totalScreens = filterLists.length > 0 ? 4 : 3;

  useEffect(() => {
    if (currentScreen === totalScreens && proxyState === 'on') {
      onClose();
    }
  }, [proxyState, currentScreen]);

  const renderCurrentScreen = () => {
    if (filterLists.length > 0) {
      switch (currentScreen) {
        case 1:
          return <WelcomeScreen />;
        case 2:
          return <FilterListsScreen filterLists={filterLists} />;
        case 3:
          return <SettingsScreen />;
        case 4:
          return <ConnectScreen />;
        default:
          return null;
      }
    }

    switch (currentScreen) {
      case 1:
        return <WelcomeScreen />;
      case 2:
        return <SettingsScreen />;
      case 3:
        return <ConnectScreen />;
      default:
        return null;
    }
  };

  return (
    <div className={`intro-fullscreen${effectiveTheme === ThemeType.DARK ? ' bp5-dark' : ''}`}>
      <AppHeader />
      <div className="intro-main-content">{renderCurrentScreen()}</div>
      <div className="intro-footer">
        {currentScreen < totalScreens ? (
          <>
            <ProgressBar
              value={currentScreen / totalScreens}
              animate={false}
              stripes={false}
              intent="primary"
              className="intro-progress-bar"
            />

            <ButtonGroup fill size="large">
              <Button fill variant="outlined" onClick={onClose} className="skip-button">
                {t('introOverlay.buttons.skip')}
              </Button>
              <Button
                fill
                intent="primary"
                onClick={() => {
                  setCurrentScreen((currentScreen) => currentScreen + 1);
                }}
                endIcon="arrow-right"
              >
                {t('introOverlay.buttons.next')}
              </Button>
            </ButtonGroup>
          </>
        ) : (
          <StartStopButton />
        )}
      </div>
    </div>
  );
}
