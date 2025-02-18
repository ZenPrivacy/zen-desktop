import { isProxyable } from './helpers/isProxyable';
import { createLogger } from './helpers/logger';
import { matchStack } from './helpers/matchStack';
import { parseRegexpFromString, parseRegexpLiteral } from './helpers/parseRegexp';

const logger = createLogger('set-constant');

export function setConstant(
  property: string,
  value: string,
  stack?: string,
  valueWrapper?: string,
  setProxyTrap?: boolean,
) {
  if (setProxyTrap !== undefined) {
    logger.warn('setProxyTrap will be ignored');
  }

  let fakeValue: any;
  switch (value) {
    case 'undefined':
      fakeValue = undefined;
      break;
    case 'false':
      fakeValue = false;
      break;
    case 'true':
      fakeValue = true;
      break;
    case 'null':
      fakeValue = null;
      break;
    case 'emptyObj':
      fakeValue = {};
      break;
    case 'emptyArr':
      fakeValue = [];
      break;
    case 'noopFunc':
      fakeValue = () => {};
      break;
    case 'noopCallbackFunc':
      fakeValue = () => () => {};
      break;
    case 'trueFunc':
      fakeValue = () => true;
      break;
    case 'falseFunc':
      fakeValue = () => false;
      break;
    case 'throwFunc':
      fakeValue = () => {
        throw new Error();
      };
      break;
    case 'noopPromiseResolve':
      fakeValue = () => {
        return Promise.resolve(
          new Response('', {
            status: 200,
            statusText: 'OK',
          }),
        );
      };
      break;
    case 'noopPromiseReject':
      fakeValue = () => Promise.reject();
      break;
    case '':
      fakeValue = '';
      break;
    case '-1':
      fakeValue = -1;
      break;
    case 'yes':
      fakeValue = 'yes';
      break;
    case 'no':
      fakeValue = 'no';
      break;
    default: {
      const int = parseInt(value, 10);
      if (!isNaN(int) && int >= 0 && int <= 32767) {
        fakeValue = value;
        break;
      }
      throw new Error('Invalid value');
    }
  }

  const wrapped = fakeValue; // Avoid creating recursive functions by creating a temporary variable.
  switch (valueWrapper) {
    case 'asFunction':
      fakeValue = () => wrapped;
      break;
    case 'asCallback':
      fakeValue = () => () => wrapped;
      break;
    case 'asResolved':
      fakeValue = () => Promise.resolve(wrapped);
      break;
    case 'asRejected':
      fakeValue = () => Promise.reject(wrapped);
      break;
  }

  let stackRe: RegExp | null;
  if (stack !== undefined && stack !== '') {
    stackRe = parseRegexpLiteral(stack) || parseRegexpFromString(stack);
  }
  stackRe ??= null;

  if (!property.includes('.')) {
    let localValue = window[property as any];
    const odesc = Object.getOwnPropertyDescriptor(window, property);
    Object.defineProperty(window, property, {
      configurable: true,
      get: () => {
        if (stackRe !== null && !matchStack(stackRe)) {
          return typeof odesc?.get === 'function' ? odesc.get.apply(window) : localValue;
        }
        logger.debug(`Returning fake value for property window.${property}`, { value });
        return fakeValue;
      },
      set:
        typeof odesc?.set === 'function'
          ? odesc?.set.bind(window)
          : (v) => {
              localValue = v;
            },
    });
    return;
  }

  // Avoid infinite recursion in case we overwrite some sub-property of Object or Function.
  const nativeObject = Object;
  const nativeFunction = Function;
  const get = (chain: string[]) => {
    let proxyCache: { proxy: any; link: any };
    let boundFnCache: Record<any, any>;
    return (target: any, key: any) => {
      if (chain.length === 1 && chain[0] === key) {
        logger.debug(`Returning fake value for property window.${property}`, { value });
        return fakeValue;
      }
      let link = Reflect.get(target, key, target);
      const desc = nativeObject.getOwnPropertyDescriptor(target, key);
      if (desc && 'value' in desc && !desc.configurable && !desc.writable) {
        // Get should return the original value for non-configurable, non-writable data properties.
        // https://tc39.es/ecma262/multipage/ordinary-and-exotic-objects-behaviours.html#sec-proxy-object-internal-methods-and-internal-slots-get-p-receiver
        return link;
      }

      if (
        typeof link === 'function' &&
        // This checks for native functions. The regex helps avoid false positives from functions containing the string "[native code]".
        // Function.prototype.toString is used to handle edge cases where a function has its toString method overridden.
        // Credit: https://stackoverflow.com/a/6599105
        /\{\s*\[native code\]/.test(nativeFunction.prototype.toString.call(link))
      ) {
        // Native functions frequently expect to be bounded to their original, **unproxied** object.
        // See https://stackoverflow.com/a/57580096 for more details.
        // Fixes https://github.com/anfragment/zen/issues/201
        if (boundFnCache !== undefined && boundFnCache[key]) {
          // Like with proxyCache, store the bound function to ensure object equality between different access operations.
          link = boundFnCache[key];
        } else {
          link = link.bind(target);
          if (boundFnCache === undefined) {
            boundFnCache = {};
          }
          boundFnCache[key] = link;
        }
      }
      if (chain[0] !== key || !isProxyable(link) || (stackRe !== null && !matchStack(stackRe))) {
        return link;
      }

      if (proxyCache?.link === link) {
        // Ensure object equality between different access operations.
        // Fixes https://github.com/anfragment/zen/issues/224
        return proxyCache.proxy;
      }
      const proxy = new Proxy(link, {
        get: get(chain.slice(1)),
      });
      proxyCache = { link, proxy };
      return proxy;
    };
  };

  const rootChain = property.split('.');
  const rootProp = rootChain.shift() as any;
  const odesc = Object.getOwnPropertyDescriptor(window, rootProp);
  let localValue = window[rootProp];
  let proxyCache: { capturedValue: any; proxy: any };
  Object.defineProperty(window, rootProp, {
    configurable: true,
    get: () => {
      let capturedValue;
      if (typeof odesc?.get === 'function') {
        // On certain properties, Safari wants window getters to be called with "window" as "this".
        // Therefore, we apply instead of doing a regular function call.
        capturedValue = odesc.get.apply(window);
      } else {
        capturedValue = localValue;
      }

      if (!isProxyable(capturedValue) || (stackRe !== null && !matchStack(stackRe))) {
        return capturedValue;
      }
      if (proxyCache?.capturedValue === capturedValue) {
        // Ensure object equality between different access operations.
        // Fixes https://github.com/anfragment/zen/issues/224
        return proxyCache.proxy;
      }
      const proxy = new Proxy(capturedValue, {
        get: get(rootChain),
      });
      proxyCache = { capturedValue, proxy };
      return proxy;
    },
    set:
      typeof odesc?.set === 'function'
        ? odesc?.set.bind(window)
        : (v) => {
            localValue = v;
          },
  });
}
