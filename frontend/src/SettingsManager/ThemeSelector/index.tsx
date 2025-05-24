import { Radio, RadioGroup, FormGroup } from '@blueprintjs/core';
import { useTranslation } from 'react-i18next';

import { ThemeType, useTheme } from '../../common/ThemeManager';

export function ThemeSelector() {
  const { t } = useTranslation();
  const { theme, setTheme } = useTheme();

  return (
    <FormGroup>
      <RadioGroup
        label={t('settings.theme.chooseTheme')}
        onChange={(e: React.FormEvent<HTMLInputElement>) => {
          const value = e.currentTarget.value as ThemeType;
          setTheme(value);
        }}
        selectedValue={theme}
        className="theme-selector__radio-group"
      >
        <Radio label={t('settings.theme.system') as string} value={ThemeType.SYSTEM} />
        <Radio label={t('settings.theme.light') as string} value={ThemeType.LIGHT} />
        <Radio label={t('settings.theme.dark') as string} value={ThemeType.DARK} />
      </RadioGroup>
    </FormGroup>
  );
}
