import { FormGroup, TextArea, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import { GetIgnoredHosts, SetIgnoredHosts } from '../../wailsjs/go/cfg/Config';
import { useProxyState } from '../context/ProxyStateContext';

export function IgnoredHostsInput() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [state, setState] = useState({
    ignoredHosts: '',
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const ignoredHosts = await GetIgnoredHosts();
      setState({ ignoredHosts: (ignoredHosts ?? []).join('\n'), loading: false });
    })();
  }, []);

  const setIgnoredHosts = useDebouncedCallback(async (ignoredHosts: string) => {
    await SetIgnoredHosts(
      ignoredHosts
        .split('\n')
        .map((host) => host.trim())
        .filter((host) => host.length > 0),
    );
  }, 500);

  return (
    <FormGroup
      label={t('ignoredHostsInput.label')}
      labelFor="ignoredHosts"
      helperText={
        <>
          {t('ignoredHostsInput.description')}
          <br />
          {t('ignoredHostsInput.helper')}
        </>
      }
    >
      <Tooltip
        content={t('common.stopProxyToModify') as string}
        disabled={!isProxyRunning}
        placement="top"
        className="settings-manager__ignored-hosts-tooltip"
      >
        <TextArea
          id="ignoredHosts"
          placeholder="example.com"
          className="settings-manager__ignored-hosts-input"
          value={state.ignoredHosts}
          onChange={(e) => {
            const { value } = e.target;
            setState({ ...state, ignoredHosts: value });
            setIgnoredHosts(value);
          }}
          disabled={state.loading || isProxyRunning}
          autoResize
          fill
        />
      </Tooltip>
    </FormGroup>
  );
}
