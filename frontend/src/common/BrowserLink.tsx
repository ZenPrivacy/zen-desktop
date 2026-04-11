import { ReactNode } from 'react';

import { BrowserOpenURL } from 'wails/runtime/runtime';

export interface BrowserLinkProps {
  href: string;
  children?: ReactNode;
}

/**
 * An accessible link that opens a URL in the default browser via BrowserOpenURL.
 */
export function BrowserLink({ href, children }: BrowserLinkProps) {
  return (
    <a
      onClick={() => BrowserOpenURL(href)}
      tabIndex={0}
      role="button"
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          BrowserOpenURL(href);
        }
      }}
    >
      {children}
    </a>
  );
}
