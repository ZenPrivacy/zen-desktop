import { Switch, FormGroup } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { AppToaster } from '@/common/toaster';
import { IsEnabled, Enable, Disable } from 'wails/go/autostart/Manager';

export function AutostartSwitch() {
  const { t } = useTranslation();
  const [state, setState] = useState({
    enabled: false,
    loading: true,
  });

  useEffect(() => {
    (async () => {
      const enabled = await IsEnabled();
      setState({ ...state, enabled, loading: false });
    })();
  }, []);

  async function disable() {
    setState((state) => ({ ...state, loading: true }));
    try {
      await Disable();
    } catch (err) {
      AppToaster.show({
        message: t('autoStartSwitch.disableError', { error: err }),
        intent: 'danger',
      });
      setState((state) => ({ ...state, loading: false }));
      return;
    }
    setState((state) => ({ ...state, enabled: false, loading: false }));
  }

  async function enable() {
    setState((state) => ({ ...state, loading: true }));
    try {
      await Enable();
    } catch (err) {
      AppToaster.show({
        message: t('autoStartSwitch.enableError', { error: err }),
        intent: 'danger',
      });
      setState((state) => ({ ...state, loading: false }));
      return;
    }
    setState((state) => ({ ...state, enabled: true, loading: false }));
  }

  return (
    <FormGroup label={t('autoStartSwitch.label')} labelFor="autostart" helperText={t('autoStartSwitch.description')}>
      <Switch
        id="autostart"
        checked={state.enabled}
        large
        disabled={state.loading}
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
