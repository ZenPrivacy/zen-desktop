import { createLogger } from './helpers/logger';
import { parseRegexpFromString, parseRegexpLiteral } from './helpers/parseRegexp';

const logger = createLogger('prevent-set-timeout');

const funcToString = (eventHandler: EventListenerOrEventListenerObject): string => {
  try {
    if (
      typeof eventHandler === 'object' &&
      'handleEvent' in eventHandler &&
      typeof eventHandler.handleEvent === 'function'
    ) {
      return eventHandler.handleEvent.toString();
    }
    return (eventHandler as EventListener).toString();
  } catch {
    return '';
  }
};

export function preventAddEventListener(event = '', search = '') {
  if (!event && !search) return;

  const eventRe = parseRegexpLiteral(event) || parseRegexpFromString(event);
  const searchRe = parseRegexpLiteral(search) || parseRegexpFromString(search);

  const handler: ProxyHandler<any> = {
    apply(target, thisArg, args) {
      const [eventType, eventListener] = args;

      const eventMatches = !!(event && eventRe?.test(eventType));
      const searchMatches = !!(search && searchRe?.test(funcToString(eventListener)));

      let shouldBlock = false;

      if (event && !search) {
        shouldBlock = eventMatches;
      } else if (!event && search) {
        shouldBlock = searchMatches;
      } else if (event && search) {
        shouldBlock = eventMatches && searchMatches;
      }

      if (shouldBlock) {
        logger.info(`Blocked addEventListener("${eventType}", ${funcToString(eventListener)})`);
        return;
      }

      return Reflect.apply(target, thisArg, args);
    },
  };

  window.addEventListener = new Proxy(window.addEventListener, handler);
  document.addEventListener = new Proxy(document.addEventListener, handler);
  Element.prototype.addEventListener = new Proxy(window.Element.prototype.addEventListener, handler);
  EventTarget.prototype.addEventListener = new Proxy(window.EventTarget.prototype.addEventListener, handler);
}
