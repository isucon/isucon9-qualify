import React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';

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

  return (
    <CircularProgress color="primary" size={80} className={classes.progress} />
  );
};

export default LoadingComponent;
