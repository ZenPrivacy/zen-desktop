import { Button, Tooltip } from '@blueprintjs/core';
import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import './index.css';
import { GetMyRules, SetMyRules } from '../../wailsjs/go/cfg/Config';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { useProxyState } from '../context/ProxyStateContext';
import { MyRulesEditor } from './editor/Editor';

const HELP_URL = 'https://github.com/ZenPrivacy/zen-desktop/blob/master/docs/external/how-to-rules.md';

export function MyRules() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();

  const [state, setState] = useState<{ rules: string; loading: boolean }>({
    rules: '',
    loading: true,
  });

  const setFilters = useDebouncedCallback(async (rules: string) => {
    await SetMyRules(rules.replace(/\r\n/g, '\n').split('\n'));
  }, 500);

  useEffect(() => {
    (async () => {
      const filters = await GetMyRules();
      setState({ rules: filters.join('\n'), loading: false });
    })();
  }, []);

  const lines = useMemo(() => state.rules.split('\n'), [state.rules]);

  return (
    <div className="my-rules">
      <div>
        <Button
          variant="outlined"
          icon="help"
          className="my-rules__help-button"
          onClick={() => BrowserOpenURL(HELP_URL)}
        >
          {t('myRules.help')}
        </Button>
      </div>

      <Tooltip
        content={t('common.stopProxyToEditRules') as string}
        disabled={!isProxyRunning}
        placement="top"
        className="my-rules__tooltip"
      >
        <MyRulesEditor
          value={state.rules}
          placeholder={t('myRules.placeholder') as string}
          disabled={isProxyRunning}
          onChange={(next) => {
            setState((s) => ({ ...s, rules: next }));
            setFilters(next);
          }}
          lines={lines}
        />
      </Tooltip>
    </div>
  );
}
