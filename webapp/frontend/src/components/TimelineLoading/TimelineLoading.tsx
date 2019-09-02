import * as React from 'react';
import { CircularProgress, Theme } from '@material-ui/core';
import makeStyles from '@material-ui/core/styles/makeStyles';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    display: 'flex',
    justifyContent: 'center',
  },
  loader: {
    margin: theme.spacing(3),
  },
}));

const TimelineLoading: React.FC<{}> = () => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <CircularProgress className={classes.loader} color="primary" />
    </div>
  );
};

export { TimelineLoading };
