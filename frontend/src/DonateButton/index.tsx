import { Button, Icon } from '@blueprintjs/core';
import { useTranslation } from 'react-i18next';

import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

import './index.css';

const LINK = 'https://opencollective.com/zen-privacy';

export function DonateButton() {
  const { t } = useTranslation();

  return (
    <Button
      icon={<Icon icon="heart" className="donate-button__icon" />}
      variant="outlined"
      onClick={() => BrowserOpenURL(LINK)}
    >
      {t('donate')}
    </Button>
  );
}
