import { Button, Classes, FormGroup, InputGroup, Switch, Tooltip } from '@blueprintjs/core';
import { InfoSign } from '@blueprintjs/icons';
import { useRef, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';

import { AddFilterList } from '../../../wailsjs/go/cfg/Config';
import { AppToaster } from '../../common/toaster';
import { useProxyState } from '../../context/ProxyStateContext';
import { FilterListType } from '../types';
import './index.css';

export function CreateFilterList({ onAdd }: { onAdd: () => void }) {
  const { t } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const urlRef = useRef<HTMLInputElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);

  const [trusted, setTrusted] = useState(false);
  const [loading, setLoading] = useState(false);

  return (
    <div className="filter-lists__create-filter-list">
      <FormGroup label="URL" labelFor="url" labelInfo="(required)">
        <InputGroup id="url" placeholder="https://example.com/filter-list.txt" required type="url" inputRef={urlRef} />
      </FormGroup>

      <FormGroup label="Name" labelFor="name" labelInfo="(optional)">
        <InputGroup id="name" placeholder="Example filter list" type="text" inputRef={nameRef} />
      </FormGroup>

      <FormGroup
        label={
          <Tooltip
            content={
              <span className={Classes.TEXT_SMALL}>
                <Trans
                  i18nKey="createFilterList.trustedListsWarning"
                  components={{
                    code: <code />,
                    strong: <strong />,
                  }}
                />
              </span>
            }
            placement="top"
            minimal
            matchTargetWidth
          >
            <span className="create-filter-list__trusted-label">
              <span>{t('filterLists.trusted')}</span>
              <InfoSign className={Classes.TEXT_MUTED} size={12} />
            </span>
          </Tooltip>
        }
        labelFor="trusted"
      >
        <Switch
          id="trusted"
          large
          checked={trusted}
          onClick={(e) => {
            setTrusted(e.currentTarget.checked);
          }}
        />
      </FormGroup>

      <Tooltip content={t('common.stopProxyToAddFilter') as string} disabled={!isProxyRunning} placement="top">
        <Button
          icon="add"
          intent="primary"
          fill
          disabled={isProxyRunning}
          onClick={async () => {
            if (!urlRef.current?.checkValidity()) {
              urlRef.current?.focus();
              return;
            }
            const url = urlRef.current?.value;
            const name = nameRef.current?.value || url;

            setLoading(true);
            const err = await AddFilterList({
              url,
              name,
              type: FilterListType.CUSTOM,
              enabled: true,
              trusted,
            });
            if (err) {
              AppToaster.show({
                message: t('createFilterList.addError', { error: err }),
                intent: 'danger',
              });
            }
            setLoading(false);
            urlRef.current!.value = '';
            nameRef.current!.value = '';
            setTrusted(false);
            onAdd();
          }}
          loading={loading}
        >
          {t('createFilterList.addList')}
        </Button>
      </Tooltip>
    </div>
  );
}
