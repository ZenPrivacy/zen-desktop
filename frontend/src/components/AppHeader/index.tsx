import { Icon, IconSize } from '@blueprintjs/core';

import { DonateButton } from '../../DonateButton';
import './index.css';

export function AppHeader() {
  return (
    <div className="heading">
      <h1 className="heading__logo">
        <Icon icon="shield" size={IconSize.LARGE} />
        ZEN
      </h1>
      <DonateButton />
    </div>
  );
}
