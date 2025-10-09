import { Trans, useTranslation } from 'react-i18next';

import { FilterLists } from '../FilterLists';
import { FilterListType } from '../FilterLists/types';

function RegionalFilterLists() {
  return <FilterLists initialType={FilterListType.REGIONAL} showTypeSelector={false} showButtons={false} />;
}

export function FilterListsScreen() {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      {/* <h2 className="intro-heading">{t('introOverlay.screen2.title')}</h2> */}
      <p className="bp5-running-text">
        <Trans
          i18nKey="introOverlay.screen2.description"
          components={{
            strong: <strong />,
          }}
        />
      </p>
      <p>{t('introOverlay.screen2.recommendation')}</p>
      <div className="filter-lists">
        <RegionalFilterLists />
      </div>
    </div>
  );
}
