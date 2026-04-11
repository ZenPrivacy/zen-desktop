import { isEmojiSupported } from 'is-emoji-supported';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import App from './App';
import { ThemeProvider } from './common/ThemeManager';
import { ProxyStateProvider } from './context/ProxyStateContext';
import ErrorBoundary from './ErrorBoundary';
import { initI18n } from './i18n';
import './styles/style.css';

(function polyfillCountryFlagEmojis() {
  if (!isEmojiSupported('😊') || isEmojiSupported('🇨🇭')) {
    return;
  }

  const style = document.createElement('style');
  style.innerHTML = `
html, body {
  font-family: 'Twemoji Country Flags', Inter, Roboto, 'Helvetica Neue', 'Arial Nova', 'Nimbus Sans', Arial, sans-serif;
}

@supports (font-variation-settings: normal) {
  html, body {
    font-family: 'Twemoji Country Flags', InterVariable, Roboto, 'Helvetica Neue', 'Arial Nova', 'Nimbus Sans', Arial, sans-serif;
  }
}
    `;
  document.head.appendChild(style);
})();

async function bootstrap() {
  await initI18n();

  const container = document.getElementById('root');
  if (!container) throw new Error('Root element not found');
  const root = createRoot(container);

  root.render(
    <StrictMode>
      <ErrorBoundary>
        <ProxyStateProvider>
          <ThemeProvider>
            <App />
          </ThemeProvider>
        </ProxyStateProvider>
      </ErrorBoundary>
    </StrictMode>,
  );
}

bootstrap();
