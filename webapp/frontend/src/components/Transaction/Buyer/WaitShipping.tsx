import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';

const useStyles = makeStyles(theme => ({}));

type Props = {};

const WaitShipping: React.FC<Props> = () => {
  const classes = useStyles();

  return (
    <React.Fragment>
      <Typography variant="h6">発送予約が完了しました</Typography>
      <Typography variant="h6">出品者からの発送をお待ちください</Typography>
    </React.Fragment>
  );
};

export default WaitShipping;
