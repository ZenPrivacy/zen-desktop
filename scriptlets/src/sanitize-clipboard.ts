import { createLogger } from './helpers/logger';

const logger = createLogger('sanitize-clipboard');

const JUNK_PARAMS = new Set<string>([
  'utm_source',
  'utm_medium',
  'utm_campaign',
  'utm_term',
  'utm_content',
  'utm_name',
  'utm_id',
  'utm_referrer',
  'fbclid',
  'gclid',
  'ref',
  'ref_src',
  'si',
  'mkt_tok',
  'igshid',
]);

function cleanText(text: string) {
  return text.replace(/https?:\/\/[^\s"']+/gi, (raw) => {
    try {
      const url = new URL(raw);
      [...url.searchParams.keys()].forEach((k) => {
        if (JUNK_PARAMS.has(k.toLowerCase())) url.searchParams.delete(k);
      });
      return url.toString();
    } catch {
      return raw;
    }
  });
}

export function sanitizeClipboard() {
  if (navigator.clipboard) {
    const handler: ProxyHandler<any> = {
      async apply(target, thisArg, args) {
        const [payload] = args;

        const txt = await Promise.resolve(payload);
        const cleaned = cleanText(String(txt));
        logger.info(`Sanitized clipboard for: ${cleaned}`);

        return Reflect.apply(target, thisArg, [cleaned]);
      },
    };

    navigator.clipboard.writeText = new Proxy(navigator.clipboard.writeText, handler);
  }

  const legacyHandler = (ev: Event): void => {
    const e = ev as ClipboardEvent | undefined;
    let text = window.getSelection()?.toString() ?? '';

    if (!text) {
      const el = document.activeElement as HTMLInputElement | HTMLTextAreaElement | null;

      if (
        el &&
        (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') &&
        el.selectionStart !== null &&
        el.selectionEnd !== null &&
        el.selectionStart !== el.selectionEnd
      ) {
        text = el.value.slice(el.selectionStart, el.selectionEnd);
      }
    }

    if (!text) return;
    const cleaned = cleanText(text);
    if (cleaned === text) return;

    if (e?.clipboardData) {
      e.clipboardData.setData('text/plain', cleaned);
      e.preventDefault();

      logger.info(`Sanitized clipboard for ${text}`);
    }
  };

  document.addEventListener('copy', legacyHandler as EventListener, true);
}
