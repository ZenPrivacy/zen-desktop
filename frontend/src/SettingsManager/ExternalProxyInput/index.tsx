import { FormGroup, HTMLSelect, InputGroup, NumericInput, Switch, Tooltip } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

import { GetExternalProxy, SetExternalProxy } from '../../../wailsjs/go/cfg/Config';
import { AppToaster } from '../../common/toaster';
import { useProxyState } from '../../context/ProxyStateContext';

interface ExternalProxyConfig {
    enabled: boolean;
    protocol: string;
    host: string;
    port: number;
    username: string;
    password: string;
}

const defaultConfig: ExternalProxyConfig = {
    enabled: false,
    protocol: 'socks5',
    host: '',
    port: 0,
    username: '',
    password: '',
};

export function ExternalProxyInput() {
    const { t } = useTranslation();
    const { isProxyRunning } = useProxyState();
    const [state, setState] = useState<ExternalProxyConfig & { loading: boolean }>({
        ...defaultConfig,
        loading: true,
    });

    useEffect(() => {
        (async () => {
            const config = await GetExternalProxy();
            setState({
                enabled: config?.enabled ?? false,
                protocol: config?.protocol ?? 'socks5',
                host: config?.host ?? '',
                port: config?.port ?? 0,
                username: config?.username ?? '',
                password: config?.password ?? '',
                loading: false,
            });
        })();
    }, []);

    const saveConfig = useDebouncedCallback(async (config: ExternalProxyConfig) => {
        try {
            await SetExternalProxy(config);
        } catch (ex) {
            AppToaster.show({
                message: t('externalProxy.saveError', { error: ex }),
                intent: 'danger',
            });
        }
    }, 500);

    const update = (patch: Partial<ExternalProxyConfig>) => {
        setState((prev) => {
            const updated = { ...prev, ...patch };
            saveConfig({
                enabled: updated.enabled,
                protocol: updated.protocol,
                host: updated.host,
                port: updated.port,
                username: updated.username,
                password: updated.password,
            });
            return updated;
        });
    };

    return (
        <FormGroup
            label={t('externalProxy.label')}
            helperText={t('externalProxy.description')}
        >
            <Tooltip
                content={t('common.stopProxyToModify') as string}
                disabled={!isProxyRunning}
                placement="top"
            >
                <Switch
                    id="externalProxy"
                    checked={state.enabled}
                    label={t('externalProxy.enabled') as string}
                    size="large"
                    disabled={state.loading || isProxyRunning}
                    onClick={() => update({ enabled: !state.enabled })}
                />
            </Tooltip>

            {state.enabled && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '8px' }}>
                    <FormGroup label={t('externalProxy.protocol')} labelFor="externalProxyProtocol" inline>
                        <HTMLSelect
                            id="externalProxyProtocol"
                            value={state.protocol}
                            options={[
                                { value: 'socks5', label: 'SOCKS5' },
                                { value: 'http', label: 'HTTP' },
                            ]}
                            onChange={(e) => update({ protocol: e.target.value })}
                            disabled={state.loading || isProxyRunning}
                        />
                    </FormGroup>

                    <FormGroup label={t('externalProxy.host')} labelFor="externalProxyHost" inline>
                        <InputGroup
                            id="externalProxyHost"
                            placeholder="127.0.0.1"
                            value={state.host}
                            onChange={(e) => update({ host: e.target.value })}
                            disabled={state.loading || isProxyRunning}
                        />
                    </FormGroup>

                    <FormGroup label={t('externalProxy.port')} labelFor="externalProxyPort" inline>
                        <NumericInput
                            id="externalProxyPort"
                            min={1}
                            max={65535}
                            value={state.port}
                            onValueChange={(port) => {
                                if (!Number.isNaN(port)) {
                                    update({ port });
                                }
                            }}
                            disabled={state.loading || isProxyRunning}
                        />
                    </FormGroup>

                    <FormGroup label={t('externalProxy.username')} labelFor="externalProxyUsername" inline>
                        <InputGroup
                            id="externalProxyUsername"
                            placeholder={t('externalProxy.optionalPlaceholder') as string}
                            value={state.username}
                            onChange={(e) => update({ username: e.target.value })}
                            disabled={state.loading || isProxyRunning}
                        />
                    </FormGroup>

                    <FormGroup label={t('externalProxy.password')} labelFor="externalProxyPassword" inline>
                        <InputGroup
                            id="externalProxyPassword"
                            type="password"
                            placeholder={t('externalProxy.optionalPlaceholder') as string}
                            value={state.password}
                            onChange={(e) => update({ password: e.target.value })}
                            disabled={state.loading || isProxyRunning}
                        />
                    </FormGroup>
                </div>
            )}
        </FormGroup>
    );
}
