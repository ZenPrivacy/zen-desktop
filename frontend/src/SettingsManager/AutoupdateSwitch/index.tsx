import { FormGroup, Switch } from '@blueprintjs/core';
import { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { GetUpdatePolicy, SetUpdatePolicy } from '../../../wailsjs/go/cfg/Config';
import { cfg } from '../../../wailsjs/go/models';
import { AppToaster } from '../../common/toaster';

export function AutoupdateSwitch() {
  const { t } = useTranslation();
  const [state, setState] = useState({
    enabled: false,
  });

  useEffect(() => {
    (async () => {
      const policy = await GetUpdatePolicy();
      setState({
        enabled: policy === cfg.UpdatePolicyType.AUTOMATIC,
      });
    })();
  }, []);

  const disable = useCallback(() => {
    (async () => {
      setState((state) => ({ ...state, loading: true }));
      try {
        await SetUpdatePolicy(cfg.UpdatePolicyType.DISABLED);
      } catch (err) {
        AppToaster.show({
          message: t('settings.updates.disableError', { error: err }),
          intent: 'danger',
        });
        setState((state) => ({ ...state, loading: false }));
        return;
      }
      setState((state) => ({ ...state, enabled: false, loading: false }));
    })();
  }, []);
  const enable = useCallback(() => {
    (async () => {
      setState((state) => ({ ...state, loading: true }));
      try {
        await SetUpdatePolicy(cfg.UpdatePolicyType.AUTOMATIC);
      } catch (err) {
        AppToaster.show({
          message: t('settings.updates.enableError', { error: err }),
          intent: 'danger',
        });
        setState((state) => ({ ...state, loading: false }));
        return;
      }
      setState((state) => ({ ...state, enabled: true, loading: false }));
    })();
  }, []);

  return (
    <FormGroup label={t('settings.updates.automaticUpdates')} helperText={t('settings.updates.description')}>
      <Switch
        id="automaticUpdates"
        checked={state.enabled}
        size="large"
        onClick={() => {
          if (state.enabled) {
            disable();
          } else {
            enable();
          }
        }}
      />
    </FormGroup>
  );
}
