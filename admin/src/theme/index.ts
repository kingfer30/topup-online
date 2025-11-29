import type { GlobalThemeOverrides } from 'naive-ui'

export const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#ff5758',
    primaryColorHover: '#ff3031',
    primaryColorPressed: '#e64546',
    primaryColorSuppl: '#ff6869',
  },
  Button: {
    colorPrimary: '#ff5758',
    colorHoverPrimary: '#ff3031',
    colorPressedPrimary: '#e64546',
    colorFocusPrimary: '#ff5758',
    borderPrimary: '1px solid #ff5758',
    borderHoverPrimary: '1px solid #ff3031',
    borderPressedPrimary: '1px solid #e64546',
    borderFocusPrimary: '1px solid #ff5758',
    textColorPrimary: '#ffffff',
    textColorHoverPrimary: '#ffffff',
    textColorPressedPrimary: '#ffffff',
    textColorFocusPrimary: '#ffffff',
  },
  Checkbox: {
    colorChecked: '#ff5758',
    checkMarkColor: '#ffffff',
    borderChecked: '1px solid #ff5758',
    borderFocus: '1px solid #ff5758',
    boxShadowFocus: '0 0 0 2px rgba(255, 87, 88, 0.2)',
  },
  Radio: {
    dotColorActive: '#ff5758',
    buttonColorActive: '#ff5758',
    boxShadowFocus: '0 0 0 2px rgba(255, 87, 88, 0.2)',
  },
  Switch: {
    railColorActive: '#ff5758',
    loadingColor: '#ff5758',
  },
  Input: {
    borderHover: '1px solid rgba(255, 87, 88, 0.5)',
    borderFocus: '1px solid #ff5758',
    boxShadowFocus: '0 0 0 2px rgba(255, 87, 88, 0.2)',
  },
  Select: {
    peers: {
      InternalSelection: {
        borderHover: '1px solid rgba(255, 87, 88, 0.5)',
        borderFocus: '1px solid #ff5758',
        boxShadowFocus: '0 0 0 2px rgba(255, 87, 88, 0.2)',
      },
    },
  },
  Slider: {
    fillColor: '#ff5758',
    fillColorHover: '#ff3031',
    dotColor: '#ff5758',
    dotColorModal: '#ff5758',
  },
  Progress: {
    fillColor: '#ff5758',
    railColor: 'rgba(255, 87, 88, 0.2)',
  },
  Tag: {
    colorPrimary: 'rgba(255, 87, 88, 0.1)',
    colorBorderedPrimary: 'rgba(255, 87, 88, 0.1)',
    borderPrimary: '1px solid rgba(255, 87, 88, 0.5)',
    textColorPrimary: '#ff5758',
  },
  Badge: {
    colorError: '#ff5758',
  },
  Tabs: {
    tabTextColorActiveBar: '#ff5758',
    tabTextColorHoverBar: '#ff3031',
    barColor: '#ff5758',
  },
}

