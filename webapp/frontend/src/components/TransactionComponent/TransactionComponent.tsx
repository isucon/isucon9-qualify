import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../../routes/Route';
import { TransactionItem } from '../../dataObjects/item';
import CardMedia from '@material-ui/core/CardMedia/CardMedia';
import CardContent from '@material-ui/core/CardContent/CardContent';
import Typography from '@material-ui/core/Typography/Typography';
import { TransactionLabel } from '../TransactionLabel';
import { Theme } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    display: 'flex',
  },
  cardContent: {
    display: 'grid',
  },
  detail: {
    display: 'flex',
    flexDirection: 'column',
  },
  itemTitle: {
    paddingLeft: theme.spacing(1),
    paddingRight: theme.spacing(3),
    paddingBottom: theme.spacing(2),
  },
  img: {
    width: '130px',
    height: '130px',
  },
}));

interface Props {
  item: TransactionItem;
}

const TransactionComponent: React.FC<Props> = ({ item }) => {
  const classes = useStyles();
  const link =
    item.status === 'on_sale'
      ? routes.item.getPath(item.id)
      : routes.transaction.getPath(item.id);

  return (
    <Card className={classes.card}>
      <RouteLink to={link}>
        <CardMedia
          className={classes.img}
          image={item.thumbnailUrl}
          title={item.name}
        />
      </RouteLink>
      <div className={classes.detail}>
        <CardContent className={classes.cardContent}>
          <Typography className={classes.itemTitle} noWrap variant="subtitle1">
            {item.name}
          </Typography>
          <TransactionLabel itemStatus={item.status} />
        </CardContent>
      </div>
    </Card>
  );
};

export { TransactionComponent };
