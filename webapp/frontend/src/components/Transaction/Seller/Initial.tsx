import React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles((theme: Theme) => ({
  button: {
    margin: theme.spacing(1),
  },
}));

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
      <Typography variant="h6">
        下記の「集荷予約」を押し、集荷予約の手続きをしてください
      </Typography>
      <Button
        className={classes.button}
        variant="contained"
        color="primary"
        onClick={onClick}
      >
        集荷予約
      </Button>
    </React.Fragment>
  );
};

export default Initial;
