import { Button, ButtonGroup } from '@blueprintjs/core';
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

const TOTAL_SCREENS = 4;

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

  useEffect(() => {
    if (currentScreen === TOTAL_SCREENS && proxyState === 'on') {
      onClose();
    }
  }, [proxyState, currentScreen]);

  const handleNextScreen = () => {
    if (currentScreen < TOTAL_SCREENS) {
      if (currentScreen === 1 && filterLists.length === 0) {
        setCurrentScreen(3);
      } else {
        setCurrentScreen(currentScreen + 1);
      }
    }
  };

  return (
    <div className={`intro-fullscreen${effectiveTheme === ThemeType.DARK ? ' bp5-dark' : ''}`}>
      <AppHeader />
      <div className="intro-main-content">
        {currentScreen === 1 ? (
          <WelcomeScreen />
        ) : currentScreen === 2 ? (
          <FilterListsScreen filterLists={filterLists} />
        ) : currentScreen === 3 ? (
          <SettingsScreen />
        ) : currentScreen === 4 ? (
          <ConnectScreen />
        ) : null}
      </div>
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
