import { Button, FormGroup, MenuItem } from '@blueprintjs/core';
import { ItemRenderer, Select } from '@blueprintjs/select';
import { useTranslation } from 'react-i18next';

import { SUPPORTED_LANGUAGES } from '../../constants/languages';

import { changeLocale, getCurrentLocale, SupportedLocale } from '../../i18n';

interface LocaleItem {
  value: SupportedLocale;
  label: string;
}

interface LocaleSelectorProps {
  showLabel?: boolean;
  showHelper?: boolean;
}

export function LocaleSelector({ showLabel = true, showHelper = true }: LocaleSelectorProps = {}) {
  const { t } = useTranslation();

  const handleLocaleChange = async (item: LocaleItem) => {
    changeLocale(item.value);
  };

  const renderItem: ItemRenderer<LocaleItem> = (item, { handleClick, handleFocus, modifiers }) => {
    return (
      <MenuItem
        active={modifiers.active}
        key={item.value}
        onClick={handleClick}
        onFocus={handleFocus}
        roleStructure="listoption"
        text={item.label}
      />
    );
  };

  const currentLocale = SUPPORTED_LANGUAGES.find((item) => item.value === getCurrentLocale()) || SUPPORTED_LANGUAGES[0];

  const selectComponent = (
    <Select<LocaleItem>
      items={SUPPORTED_LANGUAGES}
      activeItem={currentLocale}
      onItemSelect={handleLocaleChange}
      itemRenderer={renderItem}
      filterable={false}
      popoverProps={{ minimal: true }}
    >
      <Button icon="translate" text={currentLocale.label} endIcon="caret-down" />
    </Select>
  );

  if (!showLabel && !showHelper) {
    return selectComponent;
  }

  return (
    <FormGroup
      label={showLabel ? t('settings.language.label') : undefined}
      helperText={showHelper ? t('settings.language.helper') : undefined}
    >
      {selectComponent}
    </FormGroup>
  );
}
