import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const WaitShipping: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">発送予約が完了しました</Typography>
      <Typography variant="h6">出品者からの発送をお待ちください</Typography>
    </React.Fragment>
  );
};

export default WaitShipping;
