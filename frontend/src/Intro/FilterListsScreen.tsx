import { Trans, useTranslation } from 'react-i18next';

import { cfg } from '../../wailsjs/go/models';
import { FilterListItem } from '../FilterLists';

interface FilterListsScreen {
  filterLists: cfg.FilterList[];
}

export function FilterListsScreen({ filterLists }: FilterListsScreen) {
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
      <p className="bp5-running-text">{t('introOverlay.screen2.recommendation')}</p>
      <div className="filter-lists">
        {filterLists.map((l) => (
          <FilterListItem key={l.url} filterList={l} showDelete={false} showButtons={false} />
        ))}
      </div>
    </div>
  );
}
