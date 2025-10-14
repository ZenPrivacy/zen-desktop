import { Callout, Card, Divider } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { IsNoSelfUpdate } from '../../../wailsjs/go/app/App';
import { AutostartSwitch } from '../../SettingsManager/AutostartSwitch';
import { AutoupdateSwitch } from '../../SettingsManager/AutoupdateSwitch';

import './index.css';

export function SettingsScreen() {
  const { t } = useTranslation();
  const [showUpdatePolicy, setShowUpdatePolicy] = useState(false);

  useEffect(() => {
    IsNoSelfUpdate().then((noSelfUpdate) => {
      setShowUpdatePolicy(!noSelfUpdate);
    });
  }, []);

  return (
    <div className="intro-screen">
      <h3 className="bp5-heading intro-heading">{t('intro.settings.title')}</h3>
      <p className="intro-description">{t('intro.settings.description')}</p>

      <Card elevation={1} className="settings-card">
        <AutostartSwitch />

        {showUpdatePolicy && (
          <>
            <Divider className="settings-divider" />
            <AutoupdateSwitch />
          </>
        )}
      </Card>

      <Callout icon="info-sign" intent="primary" className="settings-note">
        {t('intro.settings.settingsNote')}
      </Callout>
    </div>
  );
}
