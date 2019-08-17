import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles(theme => ({}));

type Props = {
  itemId: number;
  postShipped: (itemId: number) => void;
};

const Initial: React.FC<Props> = ({ itemId, postShipped }) => {
  const classes = useStyles();

  const onClick = (e: React.MouseEvent) => {
    postShipped(itemId);
  };

  return (
    <React.Fragment>
      <Typography variant="h6">商品が購入されました</Typography>
      <Typography variant="h6">発送予約の手続きをしてください</Typography>
      <Button onClick={onClick}>発送予約</Button>
    </React.Fragment>
  );
};

export default Initial;
