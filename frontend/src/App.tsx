import { Button, ButtonGroup, Icon, IconSize, FocusStyleManager, NonIdealState } from '@blueprintjs/core';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import './App.css';

import { ThemeType, useTheme } from './common/ThemeManager';
import { useProxyState } from './context/ProxyStateContext';
import { DonateButton } from './DonateButton';
import { FilterLists } from './FilterLists';
import { MyRules } from './MyRules';
import { useProxyHotkey } from './ProxyHotkey';
import { RequestLog } from './RequestLog';
import { SettingsManager } from './SettingsManager';
import { StartStopButton } from './StartStopButton';

function App() {
  const { t } = useTranslation();
  const { effectiveTheme } = useTheme();

  useEffect(() => {
    FocusStyleManager.onlyShowFocusOnTabs();
  }, []);

  const { proxyState } = useProxyState();
  useProxyHotkey();

  const [activeTab, setActiveTab] = useState<'home' | 'filterLists' | 'myRules' | 'settings'>('home');

  return (
    <div id="app" className={effectiveTheme === ThemeType.DARK ? 'bp5-dark' : ''}>
      <div className="heading">
        <h1 className="heading__logo">
          <Icon icon="shield" size={IconSize.LARGE} />
          ZEN
        </h1>
        <DonateButton />
      </div>
      <ButtonGroup fill variant="minimal" className="tabs">
        <Button icon="circle" active={activeTab === 'home'} onClick={() => setActiveTab('home')}>
          {t('app.tabs.home')}
        </Button>
        <Button icon="filter" active={activeTab === 'filterLists'} onClick={() => setActiveTab('filterLists')}>
          {t('app.tabs.filterLists')}
        </Button>
        <Button icon="code" active={activeTab === 'myRules'} onClick={() => setActiveTab('myRules')}>
          {t('app.tabs.myRules')}
        </Button>
        <Button icon="settings" active={activeTab === 'settings'} onClick={() => setActiveTab('settings')}>
          {t('app.tabs.settings')}
        </Button>
      </ButtonGroup>

      <div className="content">
        <div style={{ display: activeTab === 'home' ? 'block' : 'none' }}>
          {proxyState === 'off' ? (
            <NonIdealState
              icon="lightning"
              title={t('app.proxy.inactive')}
              description={t('app.proxy.description') as string}
              className="request-log__non-ideal-state"
            />
          ) : (
            <RequestLog />
          )}
        </div>
        {activeTab === 'filterLists' && <FilterLists />}
        {activeTab === 'myRules' && <MyRules />}
        {activeTab === 'settings' && <SettingsManager />}
      </div>
      <StartStopButton />
    </div>
  );
}

export default App;
