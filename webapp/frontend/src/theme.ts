import { createMuiTheme } from '@material-ui/core';

export const themeInstance = createMuiTheme({
  palette: {
    background: {
      default: '#fff',
    },
    primary: {
      main: '#f44436',
    },
    secondary: {
      main: '#4fc3f7',
      contrastText: '#fff',
    },
  },
});
