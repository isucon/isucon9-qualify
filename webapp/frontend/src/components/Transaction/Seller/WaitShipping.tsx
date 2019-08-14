import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import { Typography } from '@material-ui/core';
import Button from '@material-ui/core/Button';

const useStyles = makeStyles(theme => ({}));

type Props = {
  itemId: number;
  transactionEvidenceId: number;
  postShipped: (itemId: number) => void;
};

const WaitShipping: React.FC<Props> = ({
  itemId,
  transactionEvidenceId,
  postShipped,
}) => {
  const classes = useStyles();

  const qrCodeUrl = `/transactions/${transactionEvidenceId}.png`;
  const onClick = (e: React.MouseEvent) => {
    postShipped(itemId);
  };

  return (
    <React.Fragment>
      <Typography variant="h6">購入者の発送予約が完了しました</Typography>
      <Typography variant="h6">
        配達員にこちらのQRコードを見せて発送した後、下記の発送完了を押してください
      </Typography>
      <img src={qrCodeUrl} alt="QRコード" />
      <Button onClick={onClick}>発送完了</Button>
    </React.Fragment>
  );
};

export default WaitShipping;
