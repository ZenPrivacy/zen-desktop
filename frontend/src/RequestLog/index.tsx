import { CardList, Card, Tag, Collapse, HTMLTable, Intent } from '@blueprintjs/core';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { getCurrentLocale } from '@/i18n';
import { EventsOn } from 'wails/runtime';
import './index.css';

interface Rule {
  filterName: string;
  rawRule: string;
}

interface Process {
  id: number;
  name: string;
  diskPath: string;
}

enum FilterActionKind {
  Block = 'block',
  Redirect = 'redirect',
  Modify = 'modify',
}

interface FilterAction {
  id: string;
  kind: FilterActionKind;
  method: string;
  url: string;
  to: string;
  referer: string;
  rules: Rule[];
  process: Process;
  createdAt: Date;
}

export function RequestLog() {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<FilterAction[]>([]);

  useEffect(() => {
    const cancel = EventsOn('filter:action', (action: Omit<FilterAction, 'id' | 'createdAt'>) => {
      setLogs((logs) =>
        [
          {
            ...action,
            id: id(),
            createdAt: new Date(),
          },
          ...logs,
        ].slice(0, 200),
      );
    });

    return () => {
      cancel();
    };
  }, []);

  return (
    <div className="request-log">
      {logs.length === 0 ? (
        <p className="request-log__empty">{t('requestLog.emptyState')}</p>
      ) : (
        <CardList compact>
          {logs.map((log) => (
            <RequestLogCard log={log} key={log.id} />
          ))}
        </CardList>
      )}
    </div>
  );
}

function RequestLogCard({ log }: { log: FilterAction }) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);

  const { hostname } = new URL(log.url, 'http://foo'); // Setting the base url somehow helps with parsing //hostname:port URLs

  let tagIntent: Intent;
  switch (log.kind) {
    case FilterActionKind.Block:
      tagIntent = Intent.DANGER;
      break;
    case FilterActionKind.Modify:
      tagIntent = Intent.WARNING;
      break;
    case FilterActionKind.Redirect:
      tagIntent = Intent.WARNING;
      break;
    default:
      tagIntent = Intent.NONE;
  }

  const hasProcessId = log.process.id > 0;
  const hasProcessPath = Boolean(log.process.diskPath);

  return (
    <>
      <Card key={log.id} className="request-log__card" interactive onClick={() => setIsOpen(!isOpen)}>
        <div className="request-log__card__summary">
          <Tag minimal intent={tagIntent}>
            {hostname}
          </Tag>
          <Tag minimal className="request-log__card__process-tag" title={log.process.name}>
            {log.process.name}
          </Tag>
        </div>
        <div className="bp6-text-muted">
          {log.createdAt.toLocaleTimeString(getCurrentLocale(), { timeStyle: 'short' })}
        </div>
      </Card>

      <Collapse isOpen={isOpen}>
        <Card className="request-log__card__details" compact>
          <p className="request-log__card__details__value">
            <strong>{t('requestLog.method')}: </strong>
            <Tag minimal intent="primary">
              {log.method}
            </Tag>
          </p>
          <p className="request-log__card__details__value">
            <strong>{t('requestLog.url')}: </strong>
            {log.url}
          </p>
          {log.kind === FilterActionKind.Redirect && (
            <p className="request-log__card__details__value">
              <strong>{t('requestLog.redirectedTo')}: </strong>
              {log.to}
            </p>
          )}
          {log.referer && (
            <p className="request-log__card__details__value">
              <strong>{t('requestLog.referer')}: </strong>
              {log.referer}
            </p>
          )}
          <div className="request-log__card__details__section">
            <h4 className="request-log__card__details__section-title">{t('requestLog.process')}</h4>
            <div className="request-log__card__details__process-grid">
              {hasProcessId && (
                <p className="request-log__card__details__value">
                  <strong>{t('requestLog.processId')}: </strong>
                  <Tag minimal>{log.process.id}</Tag>
                </p>
              )}
              <p className="request-log__card__details__value">
                <strong>{t('requestLog.processName')}: </strong>
                <Tag minimal className="request-log__card__details__process-name-tag" title={log.process.name}>
                  {log.process.name}
                </Tag>
              </p>
              {hasProcessPath && (
                <p className="request-log__card__details__value request-log__card__details__process-path">
                  <strong>{t('requestLog.processPath')}: </strong>
                  <span>{log.process.diskPath}</span>
                </p>
              )}
            </div>
          </div>
          <HTMLTable bordered compact striped className="request-log__card__details__rules">
            <thead>
              <tr>
                <th>{t('requestLog.filterName')}</th>
                <th>{t('requestLog.rule')}</th>
              </tr>
            </thead>
            <tbody>
              {log.rules.map((rule) => (
                <tr key={rule.rawRule}>
                  <td>{rule.filterName}</td>
                  <td>{rule.rawRule}</td>
                </tr>
              ))}
            </tbody>
          </HTMLTable>
        </Card>
      </Collapse>
    </>
  );
}

function id(): string {
  return Math.random().toString(36).slice(2, 9);
}
