import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const WaitShipping: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">商品を購入しました</Typography>
      <Typography variant="h6">出品者が発送するまでお待ち下さい</Typography>
    </React.Fragment>
  );
};

export default WaitShipping;
