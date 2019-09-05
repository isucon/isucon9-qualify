import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { ItemStatus } from '../../dataObjects/item';
import { Theme } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import Card from '@material-ui/core/Card';

const baseWidth = '80px';
const baseHeight = '25px';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    display: 'flex',
    flexDirection: 'column',
  },
  normalLabel: {
    width: baseWidth,
    height: baseHeight,
    color: theme.palette.secondary.contrastText,
    backgroundColor: theme.palette.secondary.main,
    padding: theme.spacing(1),
    textAlign: 'center',
  },
  soldOutLabel: {
    width: baseWidth,
    height: baseHeight,
    color: theme.palette.text.primary,
    backgroundColor: theme.palette.grey.A100,
    padding: theme.spacing(1),
    textAlign: 'center',
  },
  tradingLabel: {
    width: baseWidth,
    height: baseHeight,
    color: theme.palette.primary.contrastText,
    backgroundColor: theme.palette.primary.main,
    padding: theme.spacing(1),
    textAlign: 'center',
  },
}));

interface Props {
  itemStatus: ItemStatus;
}

const getLabelByStatus = (
  status: ItemStatus,
): [string, 'normalLabel' | 'soldOutLabel' | 'tradingLabel'] => {
  switch (status) {
    case 'on_sale':
      return ['販売中', 'normalLabel'];
    case 'trading':
      return ['取引中', 'tradingLabel'];
    case 'sold_out':
      return ['売り切れ', 'soldOutLabel'];
    case 'stop':
      return ['出品停止中', 'normalLabel'];
    case 'cancel':
      return ['キャンセル', 'normalLabel'];
  }
};

const TransactionLabel: React.FC<Props> = ({ itemStatus }) => {
  const classes = useStyles();
  const [labelName, classKey] = getLabelByStatus(itemStatus);
  const className = classes[classKey];

  return (
    <div className={classes.container}>
      <Card className={className}>
        <Typography variant="subtitle2" component="p">
          {labelName}
        </Typography>
      </Card>
    </div>
  );
};

export { TransactionLabel };
