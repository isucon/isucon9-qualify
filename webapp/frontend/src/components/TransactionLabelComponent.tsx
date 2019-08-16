import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import Paper from '@material-ui/core/Paper/Paper';
import { ItemStatus } from '../dataObjects/item';

const useStyles = makeStyles(theme => ({
  normalLabel: {
    width: '30px',
    height: '30px',
    color: theme.palette.secondary.light,
    backgroundColor: theme.palette.secondary.main,
  },
  soldOutLabel: {
    width: '30px',
    height: '30px',
    color: theme.palette.primary.light,
    backgroundColor: theme.palette.grey.A100,
  },
  tradingLabel: {
    width: '30px',
    height: '30px',
    color: theme.palette.primary.light,
    backgroundColor: theme.palette.primary.main,
  },
}));

type Props = {
  itemStatus: ItemStatus;
};

function getLabelByStatus(
  status: ItemStatus,
): [string, 'normalLabel' | 'soldOutLabel' | 'tradingLabel'] {
  switch (status) {
    case 'on_sale':
      return ['出品中', 'normalLabel'];
    case 'trading':
      return ['取引中', 'tradingLabel'];
    case 'sold_out':
      return ['売却済', 'soldOutLabel'];
    case 'stop':
      return ['出品停止中', 'normalLabel'];
    case 'cancel':
      return ['キャンセル', 'normalLabel'];
  }
}

const TransactionLabel: React.FC<Props> = ({ itemStatus }) => {
  const classes = useStyles();
  const [labelName, classKey] = getLabelByStatus(itemStatus);
  const className = classes[classKey];

  return <Paper className={className}>{labelName}</Paper>;
};

export { TransactionLabel };
