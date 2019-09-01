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
  postComplete: (itemId: number) => void;
};

const WaitDone: React.FC<Props> = ({ itemId, postComplete }) => {
  const classes = useStyles();

  const onClick = (e: React.MouseEvent) => {
    postComplete(itemId);
  };

  return (
    <React.Fragment>
      <Typography variant="h6">商品が発送されました</Typography>
      <Typography variant="h6">
        商品が届き次第、下記の「取引完了」を押してください
      </Typography>
      <Button
        className={classes.button}
        variant="contained"
        color="primary"
        onClick={onClick}
      >
        取引完了
      </Button>
    </React.Fragment>
  );
};

export default WaitDone;
