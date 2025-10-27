import { Trans, useTranslation } from 'react-i18next';

import { filter } from '../../../wailsjs/go/models';
import { FilterListItem } from '../../FilterLists';

interface FilterListsScreenProps {
  filterLists: filter.List[];
}

export function FilterListsScreen({ filterLists }: FilterListsScreenProps) {
  const { t } = useTranslation();

  return (
    <div className="intro-screen">
      <h3 className="bp5-heading intro-heading">{t('intro.filterLists.title')}</h3>
      <p className="bp5-running-text intro-description">
        <Trans
          i18nKey="intro.filterLists.description"
          components={{
            strong: <strong />,
          }}
        />
      </p>
      <p className="bp5-running-text intro-description">{t('intro.filterLists.recommendation')}</p>
      <div className="filter-lists">
        {filterLists.map((l) => (
          <FilterListItem key={l.url} filterList={l} showDelete={false} showButtons={false} />
        ))}
      </div>
    </div>
  );
}
