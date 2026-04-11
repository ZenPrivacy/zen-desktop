import { MenuItem } from '@blueprintjs/core';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { AppToaster } from '@/common/toaster';
import { ExportCustomFilterLists } from 'wails/go/app/App';

export function ExportFilterList() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  const handleExport = async () => {
    setLoading(true);
    try {
      await ExportCustomFilterLists();
      AppToaster.show({
        message: t('exportFilterList.successMessage'),
        intent: 'success',
      });
    } catch (error) {
      AppToaster.show({
        message: t('exportFilterList.errorMessage', { error }),
        intent: 'danger',
      });
    } finally {
      setLoading(false);
    }
  };

  return <MenuItem icon="upload" text={t('exportFilterList.export')} onClick={handleExport} disabled={loading} />;
}
