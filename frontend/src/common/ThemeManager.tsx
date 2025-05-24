import { useState, useEffect, createContext, useContext, ReactNode, useMemo } from 'react';

import { WindowSetDarkTheme, WindowSetLightTheme, WindowSetSystemDefaultTheme } from '../../wailsjs/runtime/runtime';

export enum ThemeType {
  SYSTEM = 'system',
  LIGHT = 'light',
  DARK = 'dark',
}

interface ThemeContextType {
  theme: ThemeType;
  effectiveTheme: ThemeType.DARK | ThemeType.LIGHT;
  setTheme: (theme: ThemeType) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const STORAGE_KEY = 'zen::theme';

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeType>(() => {
    const savedTheme = localStorage.getItem(STORAGE_KEY);
    return (savedTheme as ThemeType) || ThemeType.SYSTEM;
  });

  const [effectiveTheme, setEffectiveTheme] = useState<ThemeType.DARK | ThemeType.LIGHT>(ThemeType.DARK);

  const setTheme = (newTheme: ThemeType) => {
    setThemeState(newTheme);
    localStorage.setItem(STORAGE_KEY, newTheme);
    switch (newTheme) {
      case 'light':
        WindowSetLightTheme();
        break;
      case 'dark':
        WindowSetDarkTheme();
        break;
      default:
        WindowSetSystemDefaultTheme();
    }
  };

  useEffect(() => {
    if (theme !== ThemeType.SYSTEM) {
      setEffectiveTheme(theme);
      return () => {};
    }

    const syncSystemTheme = () => {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      setEffectiveTheme(prefersDark ? ThemeType.DARK : ThemeType.LIGHT);
    };

    syncSystemTheme();

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', syncSystemTheme);
    } else {
      mediaQuery.addListener(syncSystemTheme);
    }

    return () => {
      if (mediaQuery.removeEventListener) {
        mediaQuery.removeEventListener('change', syncSystemTheme);
      } else {
        mediaQuery.removeListener(syncSystemTheme);
      }
    };
  }, [theme]);

  const value = useMemo(() => ({ theme, effectiveTheme, setTheme }), [theme, effectiveTheme]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
