import { Button, TextArea, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import './index.css';
import { GetMyRules, SetMyRules } from '../../wailsjs/go/cfg/Config';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { useProxyState } from '../context/ProxyStateContext';

const HELP_URL = 'https://github.com/ZenPrivacy/zen-desktop/blob/master/docs/external/how-to-rules.md';

export function MyRules() {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [state, setState] = useState({
    rules: '',
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const filters = await GetMyRules();
      setState({ rules: filters.join('\n'), loading: false });
    })();
  }, []);

  const setFilters = useDebouncedCallback(async (rules: string) => {
    await SetMyRules(
      rules
        .split('\n')
        .map((f) => f.trim())
        .filter((f) => f.length > 0),
    );
  }, 500);

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
        <TextArea
          fill
          placeholder={t('myRules.placeholder') as string}
          className="my-rules__textarea"
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
