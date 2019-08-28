import React from 'react';
import { Typography } from '@material-ui/core';

type Props = {};

const Done: React.FC<Props> = () => {
  return (
    <React.Fragment>
      <Typography variant="h6">取引が完了しました</Typography>
    </React.Fragment>
  );
};

export default Done;
