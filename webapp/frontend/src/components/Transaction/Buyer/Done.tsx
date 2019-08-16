import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';

const useStyles = makeStyles(theme => ({}));

type Props = {};

const Done: React.FC<Props> = () => {
  const classes = useStyles();

  return (
    <React.Fragment>
      <Typography variant="h6">取引が完了しました</Typography>
    </React.Fragment>
  );
};

export default Done;
