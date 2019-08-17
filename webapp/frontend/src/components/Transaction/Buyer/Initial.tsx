import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';

const useStyles = makeStyles(theme => ({}));

type Props = {};

const Initial: React.FC<Props> = () => {
  const classes = useStyles();

  return (
    <React.Fragment>
      <Typography variant="h6">商品を購入しました</Typography>
      <Typography variant="h6">
        購入者が発送予約をするまでお待ち下さい
      </Typography>
    </React.Fragment>
  );
};

export default Initial;
