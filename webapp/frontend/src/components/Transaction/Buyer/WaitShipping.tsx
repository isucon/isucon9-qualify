import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const WaitShipping: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">商品を購入しました</Typography>
      <Typography variant="h6">商品が発送されるまでお待ち下さい</Typography>
    </React.Fragment>
  );
};

export default WaitShipping;
