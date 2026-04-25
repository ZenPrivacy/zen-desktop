import { Trans, useTranslation } from 'react-i18next';

import { FilterListItem } from '@/FilterLists';
import { config } from 'wails/go/models';

interface FilterListsScreenProps {
  filterLists: config.FilterList[];
}

export function FilterListsScreen({ filterLists }: FilterListsScreenProps) {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      <h3 className="bp6-heading intro-heading">{t('intro.filterLists.title')}</h3>
      <p className="bp6-running-text intro-description">
        <Trans
          i18nKey="intro.filterLists.description"
          components={{
            strong: <strong />,
          }}
        />
      </p>
      <p className="bp6-running-text intro-description">{t('intro.filterLists.recommendation')}</p>
      <div className="filter-lists">
        {filterLists.map((l) => (
          <FilterListItem key={l.url} filterList={l} showDelete={false} showButtons={false} />
        ))}
      </div>
    </div>
  );
}
