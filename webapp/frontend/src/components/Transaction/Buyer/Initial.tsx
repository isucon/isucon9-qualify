import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const Initial: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">商品を購入しました</Typography>
      <Typography variant="h6">購入者が発送するまでお待ち下さい</Typography>
    </React.Fragment>
  );
};

export default Initial;
