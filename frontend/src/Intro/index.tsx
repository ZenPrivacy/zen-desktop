import { Button, ButtonGroup, ProgressBar } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import './index.css';
import { GetFilterListsByLocales } from '../../wailsjs/go/cfg/Config';
import { cfg } from '../../wailsjs/go/models';
import { useProxyState } from '../context/ProxyStateContext';
import { StartStopButton } from '../StartStopButton';

import { ConnectScreen } from './ConnectScreen';
import { FilterListsScreen } from './FilterListsScreen';
import { SettingsScreen } from './SettingsScreen';
import { WelcomeScreen } from './WelcomeScreen';

interface IntroProps {
  onClose: () => void;
}

export function Intro({ onClose }: IntroProps) {
  const { t } = useTranslation();

  const [currentScreen, setCurrentScreen] = useState(1);
  const [filterLists, setFilterLists] = useState<cfg.FilterList[]>([]);
  const [filterListsLoading, setFilterListsLoading] = useState(true);

  useEffect(() => {
    GetFilterListsByLocales(navigator.languages as string[])
      .then((filterLists) => {
        if (filterLists) setFilterLists(filterLists);
        setFilterListsLoading(false);
      })
      .catch((ex) => {
        console.error(ex);
        setFilterListsLoading(false);
      });
  }, []);

  const { proxyState } = useProxyState();

  const totalScreens = filterLists.length > 0 ? 4 : 3;

  useEffect(() => {
    if (currentScreen === totalScreens && proxyState === 'on') {
      onClose();
    }
  }, [proxyState, currentScreen, totalScreens, onClose]);

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
    <>
      <div className="content">{renderCurrentScreen()}</div>
      <div className="footer">
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
                {t('intro.buttons.skip')}
              </Button>
              <Button
                fill
                intent="primary"
                onClick={() => {
                  setCurrentScreen((currentScreen) => currentScreen + 1);
                }}
                endIcon="arrow-right"
                loading={filterListsLoading}
              >
                {t('intro.buttons.next')}
              </Button>
            </ButtonGroup>
          </>
        ) : (
          <StartStopButton />
        )}
      </div>
    </>
  );
}
