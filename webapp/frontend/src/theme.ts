import { createMuiTheme } from '@material-ui/core';

const PRIMARY = '#f44436';
const SECONDARY = '#4fc3f7';
const SECONDARY_CONTRAST = '#fff';

export const themeInstance = createMuiTheme({
  palette: {
    background: {
      default: '#fff',
    },
    primary: {
      main: PRIMARY,
    },
    secondary: {
      main: SECONDARY,
      contrastText: SECONDARY_CONTRAST,
    },
  },
});
