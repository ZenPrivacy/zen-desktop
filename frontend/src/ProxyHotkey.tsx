import { useEffect } from 'react';

import { StartProxy, StopProxy } from '../wailsjs/go/app/App';

import { useProxyState } from './context/ProxyStateContext';

export function ProxyHotkey() {
  const { proxyState } = useProxyState();
  useEffect(() => {
    const spaceDown = (e: KeyboardEvent) => {
      if (e.code === 'Space' && document.activeElement === document.body) {
        if (proxyState === 'off') {
          StartProxy();
        } else if (proxyState === 'on') {
          StopProxy();
        }
      }
    };
    window.addEventListener('keydown', spaceDown);
    return () => window.removeEventListener('keydown', spaceDown);
  }, [proxyState]);
  return null;
}
