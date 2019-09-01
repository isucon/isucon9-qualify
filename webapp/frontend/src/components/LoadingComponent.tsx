import React from 'react';
import { makeStyles, MuiThemeProvider, Theme } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';
import { themeInstance } from '../theme';

const useStyles = makeStyles((theme: Theme) => ({
  progress: {
    top: 0,
    bottom: 0,
    right: 0,
    left: 0,
    margin: 'auto',
    position: 'absolute',
  },
}));

const LoadingComponent: React.FC = () => {
  const classes = useStyles();

  // MEMO: Wrap component by MuiThemeProvider again to ignore this bug. https://github.com/mui-org/material-ui/issues/14044
  return (
    <MuiThemeProvider theme={themeInstance}>
      <CircularProgress
        color="primary"
        size={80}
        className={classes.progress}
      />
    </MuiThemeProvider>
  );
};

export default LoadingComponent;
