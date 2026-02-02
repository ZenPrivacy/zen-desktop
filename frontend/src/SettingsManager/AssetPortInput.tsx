import { FormGroup, NumericInput, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import { GetAssetPort, SetAssetPort } from '../../wailsjs/go/cfg/Config';
import { AppToaster } from '../common/toaster';
import { useProxyState } from '../context/ProxyStateContext';

export function AssetPortInput() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [state, setState] = useState({
    port: 0,
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const port = await GetAssetPort();
      setState({ ...state, port, loading: false });
    })();
  }, []);

  const setPort = useDebouncedCallback(async (port: number) => {
    try {
      await SetAssetPort(port);
    } catch (ex) {
      AppToaster.show({
        message: t('assetPortInput.setError', { error: ex }),
        intent: 'danger',
      });
    }
  }, 500);

  return (
    <FormGroup
      label={t('assetPortInput.label')}
      labelFor="asset-port"
      helperText={
        <>
          {t('assetPortInput.description')}
          <br />
          {t('assetPortInput.helper')}
        </>
      }
    >
      <Tooltip content={t('common.stopProxyToModify') as string} disabled={!isProxyRunning} placement="top">
        <NumericInput
          id="asset-port"
          min={1}
          max={65535}
          value={state.port}
          onValueChange={(port) => {
            if (Number.isNaN(port)) {
              return;
            }
            setState({ ...state, port });
            setPort(port);
          }}
          disabled={state.loading || isProxyRunning}
        />
      </Tooltip>
    </FormGroup>
  );
}
