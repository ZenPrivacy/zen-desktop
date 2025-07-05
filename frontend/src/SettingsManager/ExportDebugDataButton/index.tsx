import { Button, Tooltip } from '@blueprintjs/core';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { ExportDebugData } from '../../../wailsjs/go/cfg/Config';
import { ClipboardSetText } from '../../../wailsjs/runtime';
import { AppToaster } from '../../common/toaster';

export function ExportDebugDataButton() {
  const [loading, setLoading] = useState(false);
  const { t } = useTranslation();

  return (
    <Tooltip content={t('exportDebugDataButton.tooltip') as string}>
      <Button
        loading={loading}
        onClick={async () => {
          setLoading(true);
          try {
            const debugData = await ExportDebugData();
            await ClipboardSetText(debugData);
            AppToaster.show({
              message: t('exportDebugDataButton.success'),
              intent: 'success',
            });
          } catch (err) {
            AppToaster.show({
              message: t('exportDebugDataButton.error', { error: err }),
              intent: 'danger',
            });
          } finally {
            setLoading(false);
          }
        }}
        className="export-debug-data__button"
      >
        {t('exportDebugDataButton.label')}
      </Button>
    </Tooltip>
  );
}
