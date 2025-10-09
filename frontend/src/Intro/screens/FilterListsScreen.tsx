import { useTranslation } from 'react-i18next';

import { FilterLists } from '../../FilterLists';
import { FilterListType } from '../../FilterLists/types';

function RegionalFilterLists() {
  return <FilterLists initialType={FilterListType.REGIONAL} hideTypeSelector />;
}

export function FilterListsScreen() {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      <h2 className="intro-heading">{t('introOverlay.screen2.title')}</h2>
      <p>{t('introOverlay.screen2.description')}</p>
      <p>{t('introOverlay.screen2.recommendation')}</p>
      <div className="filter-lists">
        <RegionalFilterLists />
      </div>
    </div>
  );
}
