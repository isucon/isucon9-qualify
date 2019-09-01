import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const WaitDone: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">商品が発送されました</Typography>
      <Typography variant="h6">
        購入者が商品を受け取るのをお待ち下さい
      </Typography>
    </React.Fragment>
  );
};

export default WaitDone;
