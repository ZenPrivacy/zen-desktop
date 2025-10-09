import { Callout, Card, Divider } from '@blueprintjs/core';
import { useTranslation } from 'react-i18next';

import { IsNoSelfUpdate } from '../../../wailsjs/go/app/App';
import { AutostartSwitch } from '../../SettingsManager/AutostartSwitch';
import { UpdatePolicyRadioGroup } from '../../SettingsManager/UpdatePolicyRadioGroup';

export function SettingsScreen() {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
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
    </div>
  );
}
