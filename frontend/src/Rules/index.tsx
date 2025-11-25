import { Button, TextArea, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import './index.css';
import { GetRules, SetRules } from '../../wailsjs/go/cfg/Config';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { useProxyState } from '../context/ProxyStateContext';

const HELP_URL = 'https://github.com/ZenPrivacy/zen-desktop/blob/master/docs/external/how-to-rules.md';

export function Rules() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [state, setState] = useState({
    rules: '',
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const filters = await GetRules();
      if (filters !== null) {
        setState({ rules: filters.join('\n'), loading: false });
      } else {
        setState({ ...state, loading: false });
      }
    })();
  }, []);

  const setFilters = useDebouncedCallback(async (rules: string) => {
    await SetRules(
      rules
        .split('\n')
        .map((f) => f.trim())
        .filter((f) => f.length > 0),
    );
  }, 500);

  return (
    <div className="rules">
      <div>
        <Button variant="outlined" icon="help" className="rules__help-button" onClick={() => BrowserOpenURL(HELP_URL)}>
          {t('rules.help')}
        </Button>
      </div>
      <Tooltip
        content={t('common.stopProxyToEditRules') as string}
        disabled={!isProxyRunning}
        placement="top"
        className="rules__tooltip"
      >
        <TextArea
          fill
          placeholder={t('rules.placeholder') as string}
          className="rules__textarea"
          value={state.rules}
          disabled={isProxyRunning}
          onChange={(e) => {
            const { value } = e.target;
            setState({ ...state, rules: value });
            setFilters(value);
          }}
        />
      </Tooltip>
    </div>
  );
}
