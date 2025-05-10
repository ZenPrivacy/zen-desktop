import { FormGroup, NumericInput, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import { GetPort, SetPort } from '../../wailsjs/go/cfg/Config';
import { useProxyState } from '../context/ProxyStateContext';

export function PortInput() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [state, setState] = useState({
    port: 0,
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const port = await GetPort();
      setState({ ...state, port, loading: false });
    })();
  }, []);

  const setPort = useDebouncedCallback(async (port: number) => {
    await SetPort(port);
  }, 500);

  return (
    <FormGroup
      label={t('portInput.label')}
      labelFor="port"
      helperText={
        <>
          {t('portInput.description')}
          <br />
          {t('portInput.helper')}
        </>
      }
    >
      <Tooltip content={t('common.stopProxyToModify') as string} disabled={!isProxyRunning} placement="top">
        <NumericInput
          id="port"
          min={0}
          max={65535}
          value={state.port}
          onValueChange={(port) => {
            setState({ ...state, port });
            setPort(port);
          }}
          disabled={state.loading || isProxyRunning}
        />
      </Tooltip>
    </FormGroup>
  );
}
