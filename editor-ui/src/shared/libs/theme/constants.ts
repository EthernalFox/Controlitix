import { ThemeColors } from "./types";

export const LIGHT_PALETTE: ThemeColors = {
  primary: [
    "#E7F5FF", // 0 - Lightest
    "#D0EBFF", // 1
    "#A5D8FF", // 2
    "#74C0FC", // 3
    "#4DABF7", // 4
    "#339AF0", // 5 - Основной тон
    "#228BE6", // 6 - Для кнопок/ссылок
    "#1C7ED6", // 7
    "#1971C2", // 8
    "#1864AB" // 9 - Darkest
  ],

  secondary: [
    "#EEF3FF", // 0
    "#DCE4F5", // 1
    "#B9C7E2", // 2
    "#94A8D0", // 3
    "#748DC1", // 4
    "#5F7CB8", // 5 - Основной тон
    "#5474B4", // 6
    "#44639F", // 7
    "#39588F", // 8
    "#2D4B81" // 9
  ],

  light: [
    "#F8F9FA", // 0 - Самые светлые фоны
    "#F1F3F5", // 1 - Subtle backgrounds
    "#E9ECEF", // 2 - Hover states
    "#DEE2E6", // 3 - Borders
    "#CED4DA", // 4 - Disabled elements
    "#ADB5BD", // 5 - Muted text
    "#868E96", // 6 - Secondary text
    "#495057", // 7 - Primary text
    "#343A40", // 8 - Headings
    "#212529" // 9 - Darkest text
  ]
};

export const DARK_PALETTE: ThemeColors = {
  primary: [
    "#3A9FD0", // 0 - Приглушённый accent
    "#2E8AB8", // 1
    "#2475A0", // 2
    "#1D6088", // 3
    "#174B70", // 4
    "#133D5C", // 5 - Основной тон
    "#0F2F48", // 6 - Для кнопок/ссылок
    "#0B2334", // 7
    "#081720", // 8
    "#050F14" // 9 - Darkest
  ],
  secondary: [
    "#A5B4C8", // 0
    "#8A9AAF", // 1
    "#708096", // 2
    "#5D6A7D", // 3
    "#4A5464", // 4
    "#3C4452", // 5 - Основной тон
    "#2F3540", // 6
    "#22262E", // 7
    "#17191D", // 8
    "#0C0D0F" // 9
  ],

  dark: [
    "#B0B0B0", // 0 - Lightest text (приглушённый)
    "#959595", // 1 - Light text
    "#7A7A7A", // 2 - Muted text
    "#4A4A4A", // 3 - Disabled text
    "#2A2A2A", // 4 - Borders
    "#1F1F1F", // 5 - Hover backgrounds
    "#181818", // 6 - Subtle backgrounds
    "#121212", // 7 - Card backgrounds
    "#0D0D0D", // 8 - Deep backgrounds
    "#080808" // 9 - Darkest (почти чёрный)
  ]
};
