import React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';
import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';

const useStyles = makeStyles((theme: Theme) => ({
  qrCode: {
    width: '300px',
    height: '300px',
    margin: theme.spacing(1),
  },
  button: {
    margin: theme.spacing(1),
  },
}));

type Props = {
  itemId: number;
  transactionEvidenceId: number;
  postShippedDone: (itemId: number) => void;
};

const WaitShipping: React.FC<Props> = ({
  itemId,
  transactionEvidenceId,
  postShippedDone,
}) => {
  const classes = useStyles();

  const qrCodeUrl = `/transactions/${transactionEvidenceId}.png`;
  const onClick = (e: React.MouseEvent) => {
    postShippedDone(itemId);
  };

  return (
    <Grid container>
      <Grid item xs={12}>
        <Typography variant="h6">購入者の発送予約が完了しました</Typography>
        <Typography variant="h6">
          配達員にこちらのQRコードを見せて発送した後、下記の発送完了を押してください
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <img className={classes.qrCode} src={qrCodeUrl} alt="QRコード" />
      </Grid>
      <Grid item xs={12}>
        <Button
          className={classes.button}
          variant="contained"
          color="primary"
          onClick={onClick}
        >
          発送完了
        </Button>
      </Grid>
    </Grid>
  );
};

export default WaitShipping;
